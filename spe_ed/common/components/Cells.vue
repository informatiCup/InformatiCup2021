<template>
  <!--
SPDX-License-Identifier: Apache-2.0
Copyright 2020,2021 Philipp Naumann, Marcus Soll

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->
  <div class="cells" ref="container">
    <canvas ref="canvas" :width="canvasWidth" :height="canvasHeight"></canvas>
    <div class="controls">
      <button v-if="features.includes('log') && log" @click="onLogClick">Protokoll ({{ this.log.length }})</button>
      <button v-if="features.includes('video') && video" @click="onVideoClick">Video</button>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { cellColors } from "../constants";

export default Vue.component("sp-cells", {
  model: {
    prop: "state",
    event: "change",
  },
  props: {
    state: { type: Object, default: () => undefined },
    features: { type: Array, default: () => ["log", "video"] },
  },
  watch: {
    state() {
      this.$nextTick(this.update);
    },
  },
  methods: {
    onLogClick() {
      this.download(`spe_ed-${+new Date()}.json`, new Blob([JSON.stringify(this.log)], { type: "application/json" }));
    },
    onVideoClick() {
      this.download(`spe_ed-${+new Date()}.webm`, this.video);
    },
    download(fileName, blob) {
      let a = document.createElement("a");
      a.download = fileName;
      a.href = URL.createObjectURL(blob);
      a.click();
    },
    update() {
      if (!this.state) {
        if (this.features.includes("log")) {
          this.log = undefined;
        }
        this.canvasWidth = 0;
        this.canvasHeight = 0;
        this.recorder = undefined;
        if (this.features.includes("video")) {
          this.video = undefined;
        }
        return;
      }

      if (this.features.includes("log")) {
        if (!this.log) {
          this.log = [];
        }
        this.log.push(this.state);
      }

      if (this.features.includes("video")) {
        if (window.MediaRecorder) {
          if (!this.recorder) {
            this.recorder = new MediaRecorder(this.$refs.canvas.captureStream());
            this.recorder.addEventListener("dataavailable", (event) => (this.video = event.data));
            this.recorder.start();
          }

          if (!this.state.running && !this.video) {
            this.recorder.stop();
          }
        }
      }

      let container = this.$refs.container;
      let containerWidth = container.clientWidth;
      let containerHeight = container.clientHeight;
      let base = Math.min(containerWidth, containerHeight);
      let ratio = this.state.width / this.state.height;
      if (this.state.width > this.state.height) {
        this.canvasWidth = Math.min(containerWidth, base * ratio);
        this.canvasHeight = Math.min(containerWidth / ratio, base);
      } else {
        this.canvasWidth = Math.min(containerHeight * ratio, base);
        this.canvasHeight = Math.min(containerHeight, base / ratio);
      }

      this.$nextTick(this.draw);
    },
    draw() {
      if (this.canvasWidth === 0) {
        return;
      }

      let blockSize = 20;
      let buffer = document.createElement("canvas");
      buffer.width = this.state.width * blockSize;
      buffer.height = this.state.height * blockSize;
      let bufferContext = buffer.getContext("2d");
      bufferContext.clearRect(0, 0, this.state.width * blockSize, this.state.width * blockSize);
      for (let y = 0; y < this.state.height; y++) {
        for (let x = 0; x < this.state.width; x++) {
          let value = this.state.cells[y][x];
          bufferContext.fillStyle = cellColors[value];
          bufferContext.fillRect(x * blockSize, y * blockSize, (x + 1) * blockSize, (y + 1) * blockSize);
        }
      }
      for (let player of Object.values(this.state.players)) {
        bufferContext.beginPath();
        let arcX = player.x * blockSize + blockSize / 2;
        let arcY = player.y * blockSize + blockSize / 2;
        bufferContext.arc(arcX, arcY, blockSize / 4, 0, 2 * Math.PI);
        bufferContext.fillStyle = "#ffffff";
        bufferContext.fill();
      }
      this.context.drawImage(buffer, 0, 0, this.canvasWidth, this.canvasHeight);
    },
  },
  created() {
    window.addEventListener("resize", this.update);
  },
  mounted() {
    this.context = this.$refs.canvas.getContext("2d");
  },
  beforeDestroy() {
    window.removeEventListener("resize", this.update);
  },
  data() {
    return {
      log: undefined,
      canvasWidth: 0,
      canvasHeight: 0,
      context: undefined,
      recorder: undefined,
      video: undefined,
    };
  },
});
</script>

<style lang="scss" scoped>
@import "../design/mixins";

div.cells {
  @include box-shadow;

  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  border: 1px solid #5d61a2;

  div.controls {
    z-index: 2;
    position: absolute;
    bottom: 20px;
    right: 20px;

    button {
      margin-right: 5px;

      &:last-of-type {
        margin-right: 0;
      }
    }
  }
}
</style>
