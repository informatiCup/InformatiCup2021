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
  <div class="box">
    <div class="header"><slot name="header"></slot></div>
    <div class="contents" :class="[padding && 'padding']">
      <slot name="contents"></slot>
    </div>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.component("sp-box", {
  props: {
    padding: { type: Boolean, default: false },
  },
});
</script>

<style lang="scss" scoped>
@import "../design/mixins";

div.box {
  @include box-shadow;

  border: 1px solid #5d61a2;
  margin-bottom: 20px;

  div.header {
    padding: 5px;
    color: #5d61a2;
    background: #5d61a2;
    color: #ffffff;
    text-align: center;
  }

  div.contents {
    &.padding {
      padding: 5px;
    }
  }
}
</style>
