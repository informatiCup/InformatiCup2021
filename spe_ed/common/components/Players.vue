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
  <div class="players">
    <sp-box>
      <template #header>Spieler</template>
      <template #contents>
        <div v-for="(player, id) of (state || {}).players" :class="['player', !player.active && 'inactive']" :key="id">
          <div :style="{ background: cellColors[id] }" class="color">
            <span v-if="state.you == id">●</span>
          </div>
          <div v-if="options.names" class="state">
            {{ options.names[id] }}
          </div>
          <div v-else-if="state.running" class="state">
            ({{ player.x }}, {{ player.y }}), {{ arrow(player.direction) }},
            {{ player.speed }}
          </div>
          <div v-else class="state" :title="player.name">
            {{ player.name }}
          </div>
        </div>
        <div v-if="!state" class="empty">Keine Spieler vorhanden.</div>
      </template>
    </sp-box>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import { cellColors } from "../constants";

export default Vue.component("sp-players", {
  model: {
    prop: "state",
    event: "change",
  },
  props: {
    state: { type: Object, default: () => undefined },
    options: { type: Object, default: () => ({}) },
  },
  methods: {
    arrow(direction) {
      switch (direction) {
        case "up":
          return "▲";
        case "down":
          return "▼";
        case "left":
          return "◄";
        case "right":
          return "►";
        default:
          return "?";
      }
    },
  },
  data() {
    return {
      cellColors,
    };
  },
});
</script>

<style lang="scss" scoped>
div.players {
  div.player {
    display: flex;
    width: 300px;
    line-height: 25px;
    border-bottom: 1px solid #5d61a2;

    &:last-of-type {
      border-bottom: 0;
    }

    &.inactive {
      background: #efeff6;
      color: #5f5f5f;
    }

    div.color {
      flex: 0 0 25px;
      text-align: center;
      color: #ffffff;
      border-right: 1px solid #5d61a2;
    }

    div.state {
      padding-left: 5px;
      overflow: hidden;
      white-space: nowrap;
    }
  }

  div.empty {
    padding: 0 5px;
    line-height: 25px;
  }
}
</style>
