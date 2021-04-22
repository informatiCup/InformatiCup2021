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
  <div class="server-time">
    <sp-box padding>
      <template #header>Serverzeit</template>
      <template #contents>
        <span v-if="busy">Wird abgerufen...</span>
        <span v-if="time" :class="[delta > 1000 && 'critical']">{{ time.toLocaleString() }} ({{ delta }} ms Abweichung)</span>
        <span v-if="error">Abruf der Serverzeit fehlgeschlagen.</span>
      </template>
    </sp-box>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.component("sp-server-time", {
  async mounted() {
    try {
      let response = await fetch("https://msoll.de/spe_ed_time");
      var json = await response.json();
      let time = new Date(json.time);
      time.setMilliseconds(time.getMilliseconds() + json.milliseconds);
      this.time = time;
      this.delta = Math.abs(this.time.getTime() - new Date().getTime());
      this.interval = setInterval(() => {
        this.time = new Date(this.time.getTime() + 1000);
      }, 1000);
    } catch (error) {
      this.error = error;
    }
    this.busy = false;
  },
  beforeDestroy() {
    clearInterval(this.interval);
  },
  data() {
    return {
      busy: true,
      time: undefined,
      delta: undefined,
      error: undefined,
      interval: undefined,
    };
  },
});
</script>

<style lang="scss" scoped>
span.critical {
  font-weight: bold;
  color: #f00;
}
</style>
