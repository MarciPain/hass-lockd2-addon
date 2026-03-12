# Project Architecture - hass-lockd2-addon

## Current State
The project is a Home Assistant Addon that controls devices via MQTT. Recently, MQTT publishing logic similar to ESPHome, a WebSocket connection for monitoring HA `state_changed` events, and an expanded entity filter in `app.js` (handling battery/power names) have been fully implemented.

## Development Directions and Rules
- **RULE:** After every code modification, but before Git push (or delivery to testing/user), the version number MUST be increased in `lockd2/config.yaml`. Without this, the Addon will not update in Home Assistant.
- Ensuring MQTT connection stability.
- Fine-tuning Home Assistant entity discovery (Discovery).

## File List and Functions
- [lockd2/backend/api.go](./lockd2/backend/api.go): API endpoints and Ingress middleware.
- [lockd2/backend/ha_api.go](./lockd2/backend/ha_api.go): Supervisor API integration (CallHAService implementation for controlling locks/switches) and authentication.
- [lockd2/backend/ha_ws.go](./lockd2/backend/ha_ws.go): NEW: Home Assistant WebSocket client for real-time monitoring of entity `state_changed` events.
- [lockd2/backend/mqtt.go](./lockd2/backend/mqtt.go): MQTT client with MQTT publishing and command (`cmd`) reception (with Lock/Unlock, ON/OFF logic).
- [lockd2/backend/frontend/app.js](./lockd2/backend/frontend/app.js): Client-side logic of the configuration interface with the new `battery`/`cell` filtering and UI update.
- [lockd2/config.yaml](./lockd2/config.yaml): Addon configuration.
- [lockd2/run.sh](./lockd2/run.sh): Startup script with bashio.

## Related Projects
- [lockd-go2 Backend](https://github.com/MarciPain/lockd-go2): The central backend.
- [lockd2 Mobile App](https://github.com/MarciPain/lockd2): The Flutter-based client.
