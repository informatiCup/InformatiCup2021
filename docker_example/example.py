#!/usr/bin/env python3

import asyncio
import json
import os
import random
import websockets


async def play():
    url = os.environ["URL"]
    key = os.environ["KEY"]

    async with websockets.connect(f"{url}?key={key}") as websocket:
        print("Waiting for initial state...", flush=True)
        while True:
            state_json = await websocket.recv()
            state = json.loads(state_json)
            print("<", state)
            own_player = state["players"][str(state["you"])]
            if not state["running"] or not own_player["active"]:
                break
            valid_actions = ["turn_left", "turn_right", "change_nothing"]
            if own_player["speed"] < 10:
                valid_actions += ["speed_up"]
            if own_player["speed"] > 1:
                valid_actions += ["slow_down"]
            action = random.choice(valid_actions)
            print(">", action)
            action_json = json.dumps({"action": action})
            await websocket.send(action_json)


asyncio.get_event_loop().run_until_complete(play())
