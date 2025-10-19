// generate-signature.js
import crypto from "crypto";
import fs from "fs";

// Get command-line argument (path to JSON body)
const bodyPath = process.argv[2];
if (!bodyPath) {
  console.error("❌ Usage: node generate-signature.js <path-to-body.json>");
  process.exit(1);
}

// Read the JSON body file
let body;
try {
  body = fs.readFileSync(bodyPath, "utf8").trim();
} catch (err) {
  console.error(`❌ Error reading file: ${bodyPath}\n`, err.message);
  process.exit(1);
}

// Use env var or default fallback
const secret = process.env.HMAC_SECRET || "my_super_secret_key";

// Compute HMAC-SHA256
const hmac = crypto.createHmac("sha256", secret);
hmac.update(body);
const signature = hmac.digest("hex");

// Output results
console.log("✅ HMAC Signature Generated");
console.log("----------------------------");
console.log("Body File :", bodyPath);
console.log("Secret    :", secret);
console.log("Signature :", signature);
console.log("----------------------------\n");
console.log("👉 Use this header in Bruno:");
console.log(`LT-SIGNATURE: ${signature}`);
