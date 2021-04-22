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
  <div class="messages">
    <sp-box>
      <template #header>Meldungen</template>
      <template #contents>
        <div class="scroll">
          <div v-for="(message, index) of messages" :key="index" :title="message" class="message">
            {{ message }}
          </div>
        </div>
      </template>
    </sp-box>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.component("sp-messages", {
  model: {
    prop: "messages",
    event: "change",
  },
  props: {
    messages: { type: Array, default: () => [] },
  },
});
</script>

<style lang="scss" scoped>
div.messages {
  div.scroll {
    width: 300px;
    height: 200px;
    overflow-x: hidden;
    overflow-y: auto;
  }

  div.message {
    padding: 5px;
    white-space: nowrap;

    &:nth-child(odd) {
      background: #efeff6;
    }
  }
}
</style>
