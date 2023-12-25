# Use an Awair Element to turn on an Ecobee thermostat fan

Includes very limited API support for an Awair air monitor and Ecobee thermostat, with the goal of running the fan if air quality is below a certain threshold.

Set limits for CO2, VOC, or PM2.5 as measured by the air quality monitor to trigger the fan as controlled by the thermostat.

## Usage
 
1. Copy `conf.example.toml` to `conf.toml` and edit for your devices:
   1. Set air quality limits and fan hold time in minutes
   2. Set your [Awair local address](https://support.getawair.com/hc/en-us/articles/360049221014-Awair-Element-Local-API-Feature#h_01F40FBBW5323GBPV7D6XMG4J8). This address will require `libnss-mdns` installed.
   3. Set up your [Ecobee API access](https://www.ecobee.com/home/developer/api/examples/index.shtml). You'll need an application `client_id` (the API Token) and an OAuth2 refresh token.
2. `go build`
3. run `./awair-fan`. Use `-s` flag to measure air quality without triggering the thermostat.