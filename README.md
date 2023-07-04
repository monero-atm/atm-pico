# ATM Pico
This is a small ATM build that costs around 200-300 EUR. It uses the [OpenKiosk](https://openkiosk.org) protocol and daemons.

![Image of ATM Pico's complete build](https://atm.monero.is/images/atm-pico.jpg)

## Components

- [Alberici BillyOne bill acceptor](http://www.dglpro.eu/eng-componente-acceptoare-de-bancnote.html) - 150 EUR
Any bill acceptor with pulse interface should work.
- [616 coin acceptor](https://www.aliexpress.com/w/wholesale-616-coin-acceptor.html) - 18 EUR
Any coin acceptor with pulse interface should work.
- [YHDAA M800D barcode scanner](https://www.aliexpress.com/item/1005003167832725.html) - 27 EUR
Any serial barcode/QR code scanner should work.
- [7" touchscreen HDMI display](https://www.aliexpress.com/w/wholesale-7-inch-touchscreen.html) - 44 EUR
A touchscreen is required for this build but the size doesn't matter. You can make the enclosure to fit a smaller/bigger screen.
- Cables, jumpers, shrink tubes and whatever you need to connect things
- A computer with GPIO. SBCs are great for this but you can also use external GPIO boards. We used a Raspberry Pi 3b+ for our build. But since all the software is written in Go, you can use any computer including RISC-V and MIPS architectures. 

## Software requirements
- [MoneroPay](https://moneropay.eu/guide/install.html): Suggest deploying this one in the cloud.
- [pulseacceptord](https://openkiosk.org/components/money_acceptors_pulse/)
- [codescannerd](https://openkiosk.org/components/serial_code_scanner/)
-  MQTT broker, we used [Eclipse Mosquitto](https://mosquitto.org/).

## Step 1 - Wire things together
Connect your components. Take notes of the GPIO input pins and serial device names that you will be using. This part is heavily dependent on your choice of hardware but for our build we did the following:

- BillyOne _pulse_ pin GPIO23, _enable_ pin connected to GPIO18.
- Coin acceptor _coin_ (_pulse_) pin connected to GPIO27.
- Serial QR code scanner is mapped to `/dev/ttyACM0`.
- The 7-inch touchscreen HDMI display is powered on via Pi's USB port.

### Connecting the coin and bill acceptors
Both acceptor devices require 12 volts. Since Raspberry Pi and many others use 5V, you can either choose to power these seperately or you can use a step down converter to supply 5V to the Pi from the same 12V power supply. In our MoneroKon PoC build we used two different adapters: 5V one for the Pi, 12V one for the bill and coin acceptors.

![Connected components diagram](https://atm.monero.is/images/builds/pico/atmpicoacceptors.png)

**Note:** The colors of the wires in this diagram do not match the pictures.

Close up of coin acceptor pins:

![Close up of coin acceptor pins](https://atm.monero.is/images/builds/pico/coinacceptor.jpg)

The adapter connects to all pins but the `COUNTER` pin is actually not used. This pin is undocumented, nobody really knows what it does.

Close up of bill acceptor pins:

![Close up of bill acceptor pins](https://atm.monero.is/images/builds/pico/billacceptor.jpg)

The numbers in paranthesis in the drawn diagram above are as described in BillyOne's datasheet:

![BillyOne pulse interface pin datasheet](https://atm.monero.is/images/builds/pico/billyonepulse.jpg)

## Step 2 - The software
Before starting the backend program (called atm-pico), make sure the pulseacceptord, coinacceptord and the MQTT bridge is up and running else the components might miss start/stop commands. If you have both bill and coin acceptors, you will have two instances of pulseacceptord running. You don't need to have MoneroPay running while testing but keep in mind that you won't be able to transfer Monero without it and the final step in the ATM will timeout.

### Bill acceptor configuration
The following `pulseacceptord` configuration works for Alberici BillyOne:
```yaml
device:
  pulse_pin: 23
  debounce: "100ms"
  denoise: "100ms"
  timeout: "100ms"
 
# Some devices have "enable" pins to start/stop money input.
enable_pin_control: true
# Specify if enable_pin_control is enabled.
enable_pin: 18 
enabled_when_high: false

# pulses: amount (cents)
values:
  1: 500
  2: 1000
  4: 2000
  10: 5000
  20: 10000
  40: 20000

mqtt:
  brokers:
    - "mqtt://127.0.0.1:1883"
  topic: "pulseacceptord"
  client_id: "pulseacceptord-billyone"
```

### Coin acceptor configuration
This one is tricky because we used a cheap coin acceptor which is extremely sensitive to noise and not too accurate with the pulse intervals. We recommend inserting coins and checking pulse/pause widths to verify using the [`pulse-watcher`](https://openkiosk.org/components/money_acceptors_pulse/#debugging-and-finding-config-parameters-using-pulse-watcher) program.

The `pulseacceptord` configuration below works with 616 coin acceptors from China.
```yaml
device:
  pulse_pin: 27
  debounce: "102ms"
  denoise: "28ms"
  timeout: "30ms"
  
# pulses: amount (cents). Don't forget to change these values
# if they don't match your own configuration!
values:
  2: 10
  4: 20
  10: 100

mqtt:
  brokers:
    - "mqtt://127.0.0.1:1883"
  topic: "pulseacceptord"
  client_id: "pulseacceptord-coin"
```

### QR code scanner configuration:
Following `codescannerd` configuration works for our scanner:
```yaml
device:
  portname: "/dev/ttyACM0"
  bufflen: 150
  debounce: "1s"

mqtt:
  brokers:
    - "mqtt://127.0.0.1:1883"
  topic: "codescannerd"
  client_id: "codescannerd-1"
```
### Backend configuration

The backend needs to be configured first, see `config.yaml`:
```yaml
mqtt:
  brokers:
    - "mqtt://127.0.0.1:1883"
  client_id: "atm-backend"

  topics: # ATM devices' topics
    - "pulseacceptord"
    - "codescannerd"

log_format: "pretty"
log_file: "log.txt"

# Mainnet or stagenet.
mode: "mainnet"

# This is the ATM fee percentage. For example 0.5 is 0.5%
fee: 0.5

# Address of MoneroPay instance. Can be remote.
# HTTPS and basic auth URL scheme is supported.
moneropay: "http://192.168.2.206:5000"

# Update the price value this often.
price_poll_frequency: "5m"

# Currency short name to display in UI.
currency_short: "EUR"

# Some message to display on the idle view
motd: "Welcome to Monero Konferenco 2023!"

# When to return to idle view after inactivity. If the user has
# inputted address and money into the ATM already, after this much
# time monero will be sent.
state_timeout: "5m"

# After a withdrawal has been completed, return to the idle view
# after this much time.
finish_timeout: "30s"

# Rate to use if fetching it online wasn't possible.
fallback_rate: 123.45
``` 

If you don't have a coin acceptor, bill acceptor or a QR code scanner you can omit these devices topics from the `topics` field. 
## Step 3 - Start it up
The MQTT broker must be started first, followed by the component daemons and at last the daemon.

If all went well, you should have a working ATM without an enclosure already!

![ATM working without the enclosure](https://atm.monero.is/images/builds/pico/bonelessatm.mp4)

## Step 4 - Enclosure
Order 2-3 42 cm pizzas. Enjoy them. Keep the pizza boxes. Glue and cut them like this:

Front:

![Pizza child](https://atm.monero.is/images/builds/pico/oilychild.jpg)

Back:

![Pizza child behind](https://atm.monero.is/images/builds/pico/oilychildback.jpg)

## Need help?
Join our Matrix room: [#atm:kernal.eu](https://matrix.to/#/#atm:kernal.eu)
