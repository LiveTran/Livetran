package ingest

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/datarhei/gosrt"
	"github.com/vijayvenkatj/LiveTran/internal/auth"
	"github.com/vijayvenkatj/LiveTran/internal/config"
	"github.com/vijayvenkatj/LiveTran/internal/upload"
)


func SrtConnectionTask(ctx context.Context,task *Task) {

	port,err := getFreePort()
	if err != nil {
		task.UpdateStatus(StreamStopped,fmt.Sprintf("PORT error: %s",err))
		return
	}

	addr := fmt.Sprintf(":%d",port)
	listener, err := srt.Listen("srt", addr, srt.DefaultConfig())
	if err != nil {
		task.UpdateStatus(StreamStopped,fmt.Sprintf("SRT Listener error: %s",err))
		return
	}

	defer listener.Close()

	ip := GetLocalIP()

	streamkey,err := auth.GenerateStreamKey(task.Id)
	if err != nil {
		task.UpdateStatus(StreamStopped,fmt.Sprintf("StreamKey error: %s",err))
		return
	}

	url := fmt.Sprintf("srt://%s:%d?streamid=%s",ip,port,streamkey)
	task.UpdateStatus(StreamReady,fmt.Sprintf("The stream is ready! URL -> %s",url))

	accountId := config.GetEnv("R2_ACCOUNT_ID")
	accessKey := config.GetEnv("R2_ACCESS_KEY")
	secretKey := config.GetEnv("R2_SECRET_KEY")

	uploader, err := upload.CreateCloudFlareUploader(ctx,accessKey,secretKey,accountId)
	if err != nil {
		task.UpdateStatus(StreamStopped, fmt.Sprintf("Failed to initialise Uploader : %s",err))
		return
	}

	go uploader.WatchAndUpload(ctx,"output","testing")

	var wg sync.WaitGroup
	handleStream(ctx,listener,task,&wg)
	wg.Wait()

	close(task.UpdatesChan)
}


func handleStream(ctx context.Context, listener srt.Listener, task *Task,wg *sync.WaitGroup) {

	for {

		select {

		case <- ctx.Done():
			task.UpdateStatus(StreamStopped,fmt.Sprintf("Stream Stopped: %s",context.Cause(ctx)))
			return

		default:

			cancelCtx,cancel := context.WithTimeout(ctx,120*time.Second) // Adding deadline to the ctx
			
			req,err := WaitForConnection(cancelCtx,listener,task)
			cancel() // Resourse Cleanup

			if err != nil {
				task.UpdateStatus(StreamStopped,fmt.Sprintf("%s",err))
				return
			}

			conn,err := req.Accept()
			if err != nil {
				task.UpdateStatus(StreamActive,fmt.Sprintf("Accept failed : %s",err))
				continue 
			}
			task.UpdateStatus(StreamActive,"OBS connected!")

			err = ProcessStream(ctx,conn,task,wg)
			if err != nil {
				task.UpdateStatus(StreamReady,fmt.Sprintf("Processing error: %s",err))
				continue
			}

		}
	}

}

func WaitForConnection(ctx context.Context, listener srt.Listener, task *Task) (srt.ConnRequest, error) {
	type result struct {
		req srt.ConnRequest
		err error
	}

	resultCh := make(chan result, 1)

	go func() {
		req, err := listener.Accept2()
		resultCh <- result{req, err}
	}()

	/*
	PROBLEM:
		The Accept2() function is working always to catch the incoming stream.
		Suppose if i use this in this routine, then I can't control this properly because Default would block the events. 
		ie ) Flow wont reach <- ctx.Done()
	SOLN: 
		Make a seperate goroutine to check for Accept2() 
			IF error then send the error to the error buffered channel
			IF Connection is found add it to reqChan
	*/


	select {

	case <-ctx.Done():
		cause := context.Cause(ctx)
		if (cause == context.DeadlineExceeded){
			task.UpdateStatus(StreamStopped,"TIMEOUT")
			return nil,ctx.Err()
		}

		return nil, fmt.Errorf("connection canceled or user stopped the stream")

	case res := <-resultCh:

		if res.err != nil {
			return nil, res.err
		}

		streamkey := res.req.StreamId()
		ok,_ := auth.DecodeStreamKey(task.Id,streamkey)
		if !ok {
			res.req.Reject(srt.REJ_BADSECRET)
			time.Sleep(300 * time.Millisecond)
			return WaitForConnection(ctx, listener, task)
		}

		return res.req, nil
	}
}



func ProcessStream(ctx context.Context,conn srt.Conn,task *Task,wg *sync.WaitGroup) error {
	
	cmd := exec.Command("ffmpeg",
	"-f", "mpegts",
	"-i", "pipe:0",

	// Video encoding (H.264)
	"-c:v", "libx264",
	"-preset", "ultrafast",   
	"-tune", "zerolatency",       
	"-crf", "26",    // quality level (lower is better, 18–28 is sane)

	// Audio encoding (AAC)
	"-c:a", "aac",
	"-b:a", "128k",

	// HLS segmenting
	"-f", "hls",
	"-hls_time", "3",
	"-hls_list_size", "0",
	"-hls_flags", "append_list+independent_segments",
	"-hls_segment_type", "mpegts",
	"-hls_segment_filename", fmt.Sprintf("output/%s-%%03d.ts", task.Id),

	// Final playlist file
	fmt.Sprintf("output/%s.m3u8", task.Id),
)


	stdin,err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("FFmpeg stdin error: %s", err)
	} 
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return fmt.Errorf("FFmpeg start error: %s", err)
	}

	wg.Add(1)
	go func(){
		<- ctx.Done()
		defer wg.Done()
		task.UpdateStatus(StreamStopped,"User stopped the stream!")
		stdin.Close()
		conn.Close()
		_ = cmd.Process.Signal(os.Interrupt)
		cmd.Wait()
	}()

	buf := make([]byte,1316)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			stdin.Close()
			if err := cmd.Wait(); err != nil {
				return fmt.Errorf("FFmpeg exited with error: %v", err)
			}
			return fmt.Errorf("SRT read error: %v", err)
		}

		if _, err := stdin.Write(buf[:n]); err != nil {
			stdin.Close()
			if err := cmd.Wait(); err != nil {
				return fmt.Errorf("FFmpeg exited with error: %v", err)
			}
			return fmt.Errorf("FFmpeg write error: %v", err)
		}
	}
}


func GetLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return "127.0.0.1"
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return "127.0.0.1"
}

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}


/*
	TEST SCENARIOS: 
		- Connection stopped with OBS
		- Connection stopped with Stop_Stream
		- Connection timeout
		- Connection reconnect

	Current results :
		- Reconnection makes a new MP4 file as in the code 
		- Connection stopped as expected on Stop_Stream (CommandContext of exec would SIGKILL the ffmpeg so the file is corrupted. Changed it to Command)
		- Connection timeout works as expected on NO STREAMS
		- Connection timeout works as expected on OBS TIMEOUT
*/
