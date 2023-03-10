const sizes = {
  b: 1,
  kb: 1024,
  mb: 1024 * 1024,
  gb: 1024 * 1024 * 1024,
};

onmessage = function (event) {
  const size = event.data;
  const sizeNum = parseInt(size.match(/\d+/)[0]);
  const sizeUnit = size.match(/[a-zA-Z]+/)[0].toLowerCase();
  const sizeBytes = sizeNum * sizes[sizeUnit];
  const buffer = new ArrayBuffer(sizeBytes);
  const view = new Uint8Array(buffer);

  for (let i = 0; i < sizeBytes; i++) {
    view[i] = Math.floor(Math.random() * 256);
  }

  postMessage({
    file: new Blob([buffer], { type: "application/octet-stream" }),
    size: sizeBytes,
  });
};
