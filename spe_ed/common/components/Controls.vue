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
  <div class="controls">
    <sp-box padding>
      <template #header>Steuerung</template>
      <template #contents>
        <table v-if="!!connection">
          <tr>
            <td class="label">Zeit übrig:</td>
            <td>
              <span v-if="connection.deadlineSeconds !== undefined">{{ connection.deadlineSeconds }} Sekunde</span
              ><span v-if="connection.deadlineSeconds !== undefined && connection.deadlineSeconds !== 1">n</span>
              <span v-if="connection.deadlineSeconds === undefined">Unbekannt</span>
            </td>
          </tr>
        </table>

        <button :disabled="buttonsDisabled" @click="$emit('turn-left')" title="Linksdrehung">↰</button>
        <button :disabled="buttonsDisabled" @click="$emit('turn-right')" title="Rechtsdrehung">↱</button>
        <button :disabled="buttonsDisabled" @click="$emit('speed-up')" title="Beschleunigen">↟</button>
        <button :disabled="buttonsDisabled" @click="$emit('slow-down')" title="Verlangsamen">↡</button>
        <button :disabled="buttonsDisabled" @click="$emit('change-nothing')" title="Nichts ändern">×</button>
      </template>
    </sp-box>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.component("sp-controls", {
  model: {
    prop: "connection",
    event: "change",
  },
  props: {
    disabled: { type: Boolean, default: false },
    connection: { type: Object, default: () => undefined },
  },
  computed: {
    buttonsDisabled() {
      return this.disabled || !this.connection?.established || this.connection?.passive;
    },
  },
});
</script>

<style lang="scss" scoped>
div.controls {
  table {
    margin-bottom: 5px;
    width: 100%;

    td.label {
      width: 75px;
    }
  }

  button {
    line-height: 20px;
  }

  div.time-left {
    border: 1px solid #5d61a2;
    width: 100%;

    &.disabled {
      border: 1px solid #aeb0d1;
    }
  }

  button {
    margin-right: 5px;
  }
}
</style>
