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
  <sp-state></sp-state>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.component("sp-main", {});
</script>

<style lang="scss">
@import "../design/fonts";
@import "../design/mixins";

* {
  box-sizing: border-box;
  font-family: "Noto Sans Regular", sans-serif;
  font-size: 10pt;
}

html,
body {
  margin: 0;
  padding: 0;
  width: 100%;
  height: 100%;
}
</style>
