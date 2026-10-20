# Pi 5 stereo client: the proven 10-inch build

This build replaces the 13.3-inch design. The owner ruled that no part gets ordered unless it has already been shown working on a Raspberry Pi 5. A third research round found that no published build combines a touchscreen larger than 10 inches with a DAC and an SSD on a Pi 5. Above 10 inches, no display has independent Pi 5 touch confirmation. USB-touch monitors also carry an open Pi 5 fault: certain touch controllers crash the USB controller ([raspberrypi/linux #6079](https://github.com/raspberrypi/linux/issues/6079)). The official 10-inch Touch Display 2 takes its power from the GPIO header, and OMG! Ubuntu reports it "won't work with a Pi 5 stacked with add-ons, like an active cooler and M.2/SSD HAT+" ([OMG! Ubuntu](https://www.omgubuntu.co.uk/2026/07/raspberry-pi-10-inch-touch-display)).

Every part below has an independent working report on a Pi 5, or is a first-party Raspberry Pi part designed for the Pi 5. The tradeoff the owner accepted is a 10.1-inch screen and an unmeasured DAC, in exchange for that proof.

## Parts

| Part | Pick | Price (US) | Pi 5 evidence |
|---|---|---|---|
| Display | SunFounder 10.1" HDMI touchscreen, 1280x800 IPS, 10-point | $149.99 | CNX Software tested it on a Pi 5 under labwc: "The tracking of all 10 points works pretty well" ([CNX](https://www.cnx-software.com/2025/03/11/review-of-sunfounder-10-1-inch-touchscreen-display-for-sbcs-using-raspberry-pi-5-and-radxa-rock-5b/)) |
| Power | Included 12 V/5 A adapter; the display's 5.1 V/5 A USB-C output powers the Pi | included | The configuration CNX tested |
| Cooling | Raspberry Pi Active Cooler | ~$5 | First-party; CNX's test unit had one fitted |
| Storage board | Raspberry Pi M.2 HAT+ (standard) | $12 to $17 | First-party; designed to sit over the Active Cooler ([Raspberry Pi docs](https://www.raspberrypi.com/documentation/accessories/m2-hat-plus.html)) |
| SSD | Raspberry Pi SSD 1 TB | $70 to $105 | First-party, made for the M.2 HAT+ |
| DAC | Raspberry Pi DAC Pro, RCA out | ~$30 est. (£24 at The Pi Hut) | First-party. A moOde user ran this exact stack (Pi 5, Raspberry Pi M.2 HAT, Raspberry Pi DAC Pro) and confirmed "it works" in August 2025, once a proper power supply was fitted ([moOde forum](https://moodeaudio.org/forum/showthread.php?pid=65993)) |
| Arm | Kuxiu X36 tablet clamp arm (7 to 12.9 inch, 1.5 kg rated, metal arm) | est. $60 to $80 | Parkablogs reviewed it holding a roughly 1 kg tablet ([Parkablogs](https://www.parkablogs.com/content/review-kuxiu-x36-clamp-arm-big-tablets)) |
| Spare part | Pimoroni GPIO booster header | ~$5 | Only needed if the DAC does not seat (see below) |
| Keyboard, mouse, RCA cable | Any wired or Bluetooth set; shielded RCA pair | est. $60 to $90 | Generic |

The total is about $400 to $480 excluding the Pi. An 8 GB Pi 5 is the comfortable size for a desktop client.

## How it goes together

1. Fit the Active Cooler to the Pi 5.
2. Fit the M.2 HAT+ over it with its 16 mm stacking header and spacers, and the Raspberry Pi SSD in the slot.
3. Plug the DAC Pro onto the stacking header's pins.
4. Mount the Pi on the back of the display on SunFounder's standoffs, which come pre-aligned to the Pi 5's holes (the mounting area is 85 x 70 mm). Connect HDMI, USB (touch) and the display's USB-C power output to the Pi.
5. Hold the display in the arm's clamp by its long edges. The finished unit weighs about 0.8 kg: CNX weighed 728 g with a Pi 5 and Active Cooler, and the HAT, SSD and DAC add a few dozen grams. That is about half the arm's rated load. The reviewer saw a heavy tablet sag at full low extension, so keep the arm short and high.
6. Run the RCA cable from the DAC Pro to the amp.

The display has no VESA holes, which is why the arm grips its edges. SunFounder also publishes a 3D-printable stand if a fixed mount suits the shelf better.

## The remaining uncertainty, stated plainly

The moOde report confirms the M.2 HAT+ and DAC Pro stack on a Pi 5, but it does not say whether an Active Cooler was fitted. The M.2 HAT+ is designed to sit over the cooler, so the open question is only whether the DAC Pro seats fully on the pins that rise through the HAT+. If it does not, a Pimoroni booster header (a $5 part) extends the pins. Either way nothing has to be returned. The DAC Pro has no independent audio measurement. It is chosen for its proof of working, and a first-party 2 V RMS line-out DAC is very likely transparent at line level, but no one has measured it.

The DAC Pro's EEPROM normally lets the Pi detect it automatically. If it does not appear in `aplay -l`, add its overlay from Raspberry Pi's audio documentation to `config.txt` and turn the onboard audio off.

## Setup that matters

Install Raspberry Pi OS Trixie on the SSD with Imager; with no SD card inserted, the Pi 5 boots from NVMe. For multitouch, set `mouseEmulation="no"` in `~/.config/labwc/rc.xml` or flip the Control Centre's Mouse Emulation toggle; CNX needed exactly this change. For the on-screen keyboard, set On-Screen Keyboard to "Enabled always" in Raspberry Pi Configuration, since CNX found Squeekboard otherwise stayed hidden. The software stack carries over from the earlier reports: Feishin, a local Navidrome over the synced FLAC folder, raspotify for Spotify Connect, and a PipeWire `allowed-rates` drop-in for bit-perfect output.

## First-day checks

1. Touch: run `libinput debug-events` after the multitouch setting and confirm several fingers register.
2. Stack: boot from NVMe and confirm the DAC Pro in `aplay -l`.
3. Audio: play silence at listening volume with the amp's knob setting level; add a ground-loop isolator only if you hear hum.
4. Sleep: let the screen blank, wake it, and confirm touch still registers. Touch here runs over USB, separate from the HDMI picture, so the DSI wake fault from the earlier reports should not apply. If it does fail, turn blanking off.
