const worker = new Worker("generate_file.js");
const inputGroup = document.getElementById("input-group");
const buttonGo = document.getElementById("button-go");
const latency = {
  group: document.getElementById("latency-group"),
  value: document.getElementById("latency-value"),
  progressBar: document.getElementById("latency-progress-bar"),
  inputSamplingCountElement: document.getElementById(
    "latency-input-sampling-count"
  ),
};
const downstream = {
  group: document.getElementById("downstream-group"),
  value: document.getElementById("downstream-value"),
  progressBar: document.getElementById("downstream-progress-bar"),
  inputSizeElement: document.getElementById("downstream-input-size"),
  inputSizeUnitElement: document.getElementById("downstream-input-size-unit"),
};
const upstream = {
  group: document.getElementById("upstream-group"),
  value: document.getElementById("upstream-value"),
  progressBar: document.getElementById("upstream-progress-bar"),
  inputSizeElement: document.getElementById("upstream-input-size"),
  inputSizeUnitElement: document.getElementById("upstream-input-size-unit"),
};

downstream.inputSizeElement.addEventListener("input", validateNumber);
upstream.inputSizeElement.addEventListener("input", validateNumber);
latency.inputSamplingCountElement.addEventListener("input", validateNumber);

function validateNumber() {
  const inputValue = parseInt(this.value);

  if (isNaN(inputValue) || inputValue < this.min || inputValue > this.max) {
    this.classList.add("invalid");
    buttonGo.disabled = true;
  } else {
    this.classList.remove("invalid");
    buttonGo.disabled = false;
  }
}

function inputDisabled(disabled) {
  buttonGo.disabled = disabled;
  downstream.inputSizeElement.disabled = disabled;
  downstream.inputSizeUnitElement.disabled = disabled;
  upstream.inputSizeElement.disabled = disabled;
  upstream.inputSizeUnitElement.disabled = disabled;
  if (disabled) {
    inputGroup.classList.add("disabled");
    latency.group.classList.add("hidden");
    downstream.group.classList.add("hidden");
    upstream.group.classList.add("hidden");
    latency.progressBar.value = 0;
    downstream.progressBar.value = 0;
    upstream.progressBar.value = 0;
    latency.value.innerHTML = "0";
    downstream.value.innerHTML = "0";
    upstream.value.innerHTML = "0";
    latency.progressBar.classList.remove("invalid");
    downstream.progressBar.classList.remove("invalid");
    upstream.progressBar.classList.remove("invalid");
  } else {
    inputGroup.classList.remove("disabled");
  }
}

function go() {
  inputDisabled(true);
  setTimeout(measureLatency, 500);
}

function measureLatency() {
  latency.group.classList.remove("hidden");

  let latencySampling = Array();
  const samplingCount = latency.inputSamplingCountElement.value;
  const startTime = performance.now();

  const measureLatency = () => {
    fetch("/latency")
      .then(() => {
        const endTime = performance.now();
        const latencyResult = endTime - startTime;
        latency.progressBar.value =
          ((latencySampling.length + 1) / samplingCount) * 100;
        latencySampling.push(latencyResult);

        latency.value.innerHTML = (
          latencySampling.reduce(
            (accumulator, currentValue) => accumulator + currentValue
          ) / samplingCount
        ).toFixed(2);

        if (latencySampling.length < samplingCount) {
          measureLatency();
        } else {
          measureDownstream();
        }

        return null;
      })
      .catch((error) => {
        latency.progressBar.value = 100;
        latency.progressBar.classList.add("invalid");
        console.error(error);
      });
  };
  measureLatency();
}

function measureDownstream() {
  downstream.group.classList.remove("hidden");

  var xhr = new XMLHttpRequest();
  xhr.open(
    "GET",
    `/downstream?size=${downstream.inputSizeElement.value}${downstream.inputSizeUnitElement.value}`,
    true
  );
  xhr.responseType = "blob";
  xhr.onprogress = function (event) {
    if (event.lengthComputable) {
      const loaded = event.position || event.loaded;
      const total = event.totalSize || event.total;
      const speed = (
        loaded /
        1024 /
        1024 /
        ((performance.now() - startTime) / 1000)
      ).toFixed(2);

      downstream.progressBar.value = Math.round((loaded / total) * 100);
      downstream.value.innerHTML = speed;
    }
  };
  xhr.onreadystatechange = function () {
    if (xhr.readyState === XMLHttpRequest.DONE) {
      if (xhr.status === 200) {
        measureUpstream();
      } else {
        downstream.progressBar.value = 100;
        downstream.progressBar.classList.add("invalid");
        console.error("Request failed with status:", xhr.status);
      }
    }
  };
  xhr.onerror = function (error) {
    downstream.progressBar.value = 100;
    downstream.progressBar.classList.add("invalid");
    console.error(error);
  };

  const startTime = performance.now();
  xhr.send();
}

function measureUpstream() {
  upstream.group.classList.remove("hidden");
  worker.onmessage = function (event) {
    const { file: file, size: fileSize } = event.data;

    const xhr = new XMLHttpRequest();
    xhr.open("POST", "/upstream");
    const formData = new FormData();
    formData.append("file", file);

    xhr.upload.onprogress = function (e) {
      if (e.lengthComputable) {
        const loaded = e.position || e.loaded;
        const total = e.totalSize || fileSize;
        const speed = (
          loaded /
          1024 /
          1024 /
          ((performance.now() - startTime) / 1000)
        ).toFixed(2);

        upstream.progressBar.value = Math.round((loaded / total) * 100);
        upstream.value.innerHTML = speed;
      }
    };
    xhr.onerror = function (error) {
      upstream.progressBar.value = 100;
      upstream.progressBar.classList.add("invalid");
      console.error(error);
    };
    xhr.onreadystatechange = function () {
      if (xhr.readyState === 4 && xhr.status === 408) {
        upstream.progressBar.value = 100;
        upstream.progressBar.classList.add("invalid");
        console.error("Request timed out");
      }
    };
    xhr.onload = function () {
      inputDisabled(false);
    };

    const startTime = performance.now();
    xhr.send(formData);
  };

  worker.postMessage(
    `${upstream.inputSizeElement.value}${upstream.inputSizeUnitElement.value}`
  );
}
