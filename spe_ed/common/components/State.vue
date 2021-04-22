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
  <div class="state">
    <div class="left">
      <sp-cells v-model="stateInternal" :features="cellsFeatures"></sp-cells>
    </div>
    <div class="right">
      <sp-connection
        v-if="modules.includes('connection')"
        v-model="connection"
        :disabled="busy"
        @connect="onConnect"
        @disconnect="onDisconnect"
      ></sp-connection>
      <sp-server-time v-if="modules.includes('serverTime')"></sp-server-time>
      <sp-controls
        v-if="modules.includes('controls')"
        v-model="connection"
        :disabled="busy"
        @turn-left="send('turn_left')"
        @turn-right="send('turn_right')"
        @speed-up="send('speed_up')"
        @slow-down="send('slow_down')"
        @change-nothing="send('change_nothing')"
      ></sp-controls>
      <sp-players v-if="modules.includes('players')" v-model="stateInternal" :options="moduleOptions.players"></sp-players>
      <sp-messages v-if="modules.includes('messages')" v-model="messages"></sp-messages>
      <slot name="custom-modules"></slot>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";
import dayjs from "dayjs";
import WebsocketClient from "websocket-async";

export default Vue.component("sp-state", {
  model: {
    prop: "state",
    event: "changed",
  },
  props: {
    modules: {
      type: Array,
      default: () => ["connection", "controls", "players", "messages", "serverTime"],
    },
    moduleOptions: {
      type: Object,
      default: () => ({}),
    },
    state: { type: Object, default: () => undefined },
    cellsFeatures: { type: Array, default: () => ["log", "video"] },
  },
  watch: {
    state() {
      this.stateInternal = this.state;
    },
    stateInternal() {
      this.$emit("changed", this.stateInternal);
    },
  },
  methods: {
    async onDocumentKeyDown(event) {
      if (this.busy || !this.connection.established) {
        return false;
      }

      switch (event.code) {
        case "ArrowLeft":
          await this.send("turn_left");
          break;
        case "ArrowRight":
          await this.send("turn_right");
          break;
        case "ArrowUp":
          await this.send("speed_up");
          break;
        case "ArrowDown":
          await this.send("slow_down");
          break;
        case "Space":
          await this.send("change_nothing");
          break;
      }
    },
    async onConnect() {
      this.busy = true;
      this.stateInternal = undefined;
      this.log(`Verbinde mit "${this.connection.url}"...`);
      let url = `${this.connection.url}?key=${this.connection.key}`;
      this.connection.client = new WebsocketClient();
      try {
        await this.connection.client.connect(url);
        this.connection.established = true;
        this.log("Verbunden. Warte auf Initialzustand...");
        await this.receive();
      } catch {
        this.log("Verbindung fehlgeschlagen.");
      }
      this.busy = false;
    },
    async onDisconnect() {
      await this.connection.client.disconnect();
      this.connection.deadline = undefined;
      this.connection.deadlineSeconds = undefined;
      clearInterval(this.tickInterval);
      this.connection.established = false;
      this.log("Getrennt.");
    },
    async receive() {
      let json;
      try {
        json = await this.connection.client.receive();
      } catch {
        this.connection.established = false;
        this.busy = false;
        this.log("Getrennt.");
        return;
      }

      let state;
      try {
        state = JSON.parse(json);
      } catch {
        this.log("Ungültigen Zustand empfangen.");
        return;
      }
      this.stateInternal = state;
      this.connection.passive = !state.players[state.you].active;
      this.connection.deadline = dayjs(state.deadline).subtract(2, "second");

      if (!this.stateInternal.running) {
        this.connection.established = false;
        this.busy = false;
        this.log("Spiel beendet.");
        return;
      }

      if (this.connection.passive) {
        await this.receive();
      } else {
        this.busy = false;
        this.tickInterval = setInterval(this.tick, 500);
        this.tick();
      }
    },
    async send(action) {
      this.busy = true;
      this.connection.deadline = undefined;
      this.connection.deadlineSeconds = undefined;
      clearInterval(this.tickInterval);
      await this.connection.client.send(JSON.stringify({ action }));
      this.log(`Aktion "${action}" gesendet.`);
      await this.receive();
    },
    tick() {
      if (this.connection.deadline) {
        if (dayjs().isAfter(this.connection.deadline)) {
          this.send("change_nothing");
        } else {
          this.connection.deadlineSeconds = this.connection.deadline.diff(dayjs(), "second");
        }
      }
    },
    log(message) {
      this.messages.unshift(`${dayjs().format("HH:mm:ss")}: ${message}`);
    },
  },
  created() {
    document.addEventListener("keydown", this.onDocumentKeyDown);
  },
  beforeDestroy() {
    document.removeEventListener("keydown", this.onDocumentKeyDown);
  },
  data() {
    return {
      busy: false,
      connection: {
        url: "wss://msoll.de/spe_ed",
        key: "",
        client: undefined,
        established: false,
        passive: false,
        deadline: undefined,
        deadlineSeconds: undefined,
      },
      tickInterval: undefined,
      stateInternal: undefined,
      messages: [],
    };
  },
});
</script>

<style lang="scss" scoped>
div.state {
  display: flex;
  width: 100%;
  height: 100%;
}

div.left {
  flex: 1;
  padding: 20px;
  overflow: hidden;
}

div.right {
  display: flex;
  flex: 0 0 300px;
  flex-direction: column;
  padding: 20px;
  padding-left: 0;
  height: 100%;
}
</style>
