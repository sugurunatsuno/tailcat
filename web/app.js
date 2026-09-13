const params = new URLSearchParams(location.hash.slice(1));
const version = params.get("v");
const addr = params.get("tc");
document.querySelector("#status").textContent = version === "1" && addr ? "Connecting…" : "Invalid link";

