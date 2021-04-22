/*
 * SPDX-License-Identifier: Apache-2.0
 * Copyright 2020,2021 Philipp Naumann, Marcus Soll
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *	  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
import Vue from "vue";

// Components
import StateComponent from "./components/State.vue";
Vue.component(StateComponent);
import CellsComponent from "./components/Cells.vue";
Vue.component(CellsComponent);
import PlayersComponent from "./components/Players.vue";
Vue.component(PlayersComponent);
import MessagesComponent from "./components/Messages.vue";
Vue.component(MessagesComponent);
import ConnectionComponent from "./components/Connection.vue";
Vue.component(ConnectionComponent);
import ControlsComponent from "./components/Controls.vue";
Vue.component(ControlsComponent);
import BoxComponent from "./components/Box.vue";
Vue.component(BoxComponent);
import ServerTimeComponent from "./components/ServerTime.vue";
Vue.component(ServerTimeComponent);
import PlayerComponent from "./custom-components/Player.vue";
Vue.component(PlayerComponent);

import MainComponent from "./custom-components/Main.vue";
new Vue({
  render: h => h(MainComponent)
}).$mount("#main");
