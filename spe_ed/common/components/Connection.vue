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
  <div class="connection">
    <sp-box padding>
      <template #header>Verbindung</template>
      <template #contents>
        <table>
          <tr>
            <td class="label">URL:</td>
            <td>
              <input v-model="connection.url" :disabled="disabled || connection.established" name="url" />
            </td>
          </tr>
          <tr>
            <td class="label">API-Key:</td>
            <td>
              <input v-model="connection.key" :disabled="connection.established" name="key" />
            </td>
          </tr>
        </table>

        <button :disabled="disabled || !connection.key || connection.established" @click="$emit('connect')" accesskey="1">Verbinden</button>
        <button :disabled="disabled || !connection.established" @click="$emit('disconnect')" accesskey="2">Trennen</button>
      </template>
    </sp-box>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.component("sp-connection", {
  model: {
    prop: "connection",
    event: "change",
  },
  props: {
    disabled: { type: Boolean, default: false },
    connection: { type: Object, default: () => undefined },
  },
});
</script>

<style lang="scss" scoped>
div.connection {
  table {
    margin-bottom: 5px;
    width: 100%;

    td.label {
      width: 60px;
    }
  }

  button {
    margin-right: 5px;
  }

  input {
    width: 100%;
  }
}
</style>
