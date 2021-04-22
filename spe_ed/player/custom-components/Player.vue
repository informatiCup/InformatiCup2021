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
  <sp-state v-model="state" :modules="['players']" :module-options="moduleOptions" :cells-features="[]">
    <template #custom-modules>
      <sp-box padding>
        <template #header>Spiel</template>
        <template #contents>
          <table v-if="!!states">
            <tr>
              <td class="label">Runde:</td>
              <td>
                {{ round + 1 }}
              </td>
            </tr>
          </table>

          <button :disabled="!!autoplayInterval" @click="load" alt="Laden">🗀</button>
          <button :disabled="!states || round === 0" @click="first" alt="Erste Runde">⭰</button>
          <button :disabled="!states || round === 0" @click="previous" alt="Vorige Runde">⭠</button>
          <button :disabled="!states || round === states.length - 1" @click="next" alt="Nächste Runde">⭢</button>
          <button :disabled="!states || round === states.length - 1" @click="last" alt="Letzte Runde">⭲</button>
          <button :disabled="!states || round === states.length - 1" @click="toggleAutoplay" alt="Autoplay">
            <span v-if="autoplayInterval">Stop</span><span v-else>Start</span>
          </button>
          <input ref="fileInput" type="file" style="display: none" @change="onFileInputChange" />
        </template>
      </sp-box>
    </template>
  </sp-state>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.component("sp-player", {
  watch: {
    round() {
      this.state = this.states[this.round];
    }
  },
  computed: {
    moduleOptions() {
      if (!this.states || this.states.length === 0) {
        return {};
      }
      let lastState = this.states[this.states.length - 1];
      let names = {};
      for (let id in lastState.players) {
        names[id] = lastState.players[id].name;
      }
      return {
        players: { names }
      };
    }
  },
  methods: {
    onDocumentKeyDown(event) {
      switch (event.code) {
        case "Home":
          event.preventDefault();
          this.first();
          break;
        case "ArrowLeft":
          event.preventDefault();
          this.previous();
          break;
        case "ArrowRight":
          event.preventDefault();
          this.next();
          break;
        case "End":
          event.preventDefault();
          this.last();
          break;
        case "Space":
          event.preventDefault();
          this.toggleAutoplay();
          break;
      }
    },
    onFileInputChange() {
      this.readFile(this.$refs.fileInput.files[0]);
    },
    readFile(file) {
      if (!file) {
        return;
      }
      this.states = [];
      let reader = new FileReader();
      reader.addEventListener("load", event => {
        let log = event.target.result.toString().split("\n");
        this.meta = JSON.parse(log[0]);
        let states = [];
        for (let i = 1; i < log.length; i++) {
          if (log[i].length > 0) {
            states.push(JSON.parse(log[i]));
          }
        }
        this.states = states;
        this.round = 0;
        this.state = this.states[0];
      });
      reader.readAsText(file);
    },
    async load() {
      this.$refs.fileInput.click();
    },
    first() {
      this.round = 0;
    },
    previous() {
      if (this.round === 0) {
        return false;
      }
      this.round -= 1;
      return true;
    },
    next() {
      if (this.round === this.states.length - 1) {
        return false;
      }
      this.round += 1;
      return true;
    },
    last() {
      this.round = this.states.length - 1;
    },
    startAutoplay() {
      this.autoplayInterval = setInterval(() => {
        if (!this.next()) {
          this.stopAutoplay();
        }
      }, 500);
    },
    stopAutoplay() {
      clearInterval(this.autoplayInterval);
      this.autoplayInterval = undefined;
    },
    toggleAutoplay() {
      if (this.autoplayInterval) {
        this.stopAutoplay();
      } else {
        this.startAutoplay();
      }
    }
  },
  mounted() {
    document.addEventListener("keydown", this.onDocumentKeyDown);
  },
  beforeDestroy() {
    document.removeEventListener("keydown", this.onDocumentKeyDown);
  },
  data() {
    return {
      meta: undefined,
      states: undefined,
      state: undefined,
      round: undefined,
      autoplayInterval: undefined
    };
  }
});
</script>

<style lang="scss" scoped>
table {
  margin-bottom: 5px;
  width: 100%;

  td.label {
    width: 50px;
  }
}

button {
  line-height: 20px;
}
</style>
