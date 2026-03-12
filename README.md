# Lockd2 Home Assistant Addon

A Home Assistant addon for integrating the Lockd2 ecosystem. It provides a web interface (Ingress) and connects your Home Assistant entities to the Lockd2 MQTT network.

## Features
- **Ingress Support**: Configure and manage your locks directly from the Home Assistant UI.
- **MQTT Bridge**: Automatically publishes state changes to MQTT in an ESPHome-compatible format.
- **WebSocket Integration**: Listens to Home Assistant `state_changed` events in real-time.
- **Battery Monitoring**: Automatically filters and displays battery/power entities.

## Installation
1. Add this repository to your Home Assistant Addon Store:
   `https://github.com/MarciPain/hass-lockd2-addon`
2. Install the **lockd2** addon.
3. Configure the MQTT broker and other settings in the **Configuration** tab.
4. Start the addon and open the **Web UI**.

## Related Projects
- **[lockd-go2 Backend](https://github.com/MarciPain/lockd-go2)**: The core service that manages locks and security.
- **[lockd2 Mobile App](https://github.com/MarciPain/lockd2)**: Flutter application for mobile access.

## Technical Details
See [ARCHITECTURE.md](./ARCHITECTURE.md) for a detailed internal structure of the addon.

---

[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-Donate-orange.svg)](https://buymeacoffee.com/marcipain)
