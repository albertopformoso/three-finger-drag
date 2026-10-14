# Three Finger Drag

This program provides a three finger drag utility for Linux systems. The utility works on Arch Linux and CachyOS. It replicates the exact drag behavior found in macOS. Users can move windows, select text, and drag files using three fingers without pressing down on the trackpad.

## Requirements

+ Arch Linux or CachyOS.
+ Wayland or X11 display server.
+ Libinput library.
+ C compiler and Make utility.

## Installation

Follow these steps to install the program.

1. Open a terminal window.

2. Install the build dependencies:
    `sudo pacman -S base-devel libinput`

3. Clone this repository:
    `git clone [https://github.com/yourusername/linux-three-finger-drag.git](https://github.com/yourusername/linux-three-finger-drag.git)`

4. Enter the directory:
    `cd linux-three-finger-drag`

5. Compile the program:
    `make`

6. Install the binary:
    `sudo make install`

## Configuration

The configuration file is located at `~/.config/three-finger-drag/config.conf`. You can edit this file to change the drag timeout and sensitivity values.
Usage

Start the service with this command:
`three-finger-drag-start`

Enable the systemd user service to run the program at startup:
`systemctl --user enable --now three-finger-drag.service`

## Project Structure

```bash
.
├── AGENTS.md
├── docs
│   └── constitution.md
├── experiments
│   └── inputprobe
│       ├── drag.go
│       ├── drag_runtime.go
│       ├── drag_runtime_test.go
│       ├── drag_test.go
│       ├── main.go
│       ├── main_test.go
│       ├── native.c
│       ├── native.go
│       ├── native.h
│       ├── native_integration_test.go
│       ├── README.md
│       ├── steady.go
│       ├── steady_test.go
│       ├── transfer.go
│       ├── transfer_test.go
│       ├── travel.go
│       └── travel_test.go
├── go.mod
├── internal
│   ├── gesture
│   │   ├── contacts.go
│   │   ├── contacts_test.go
│   │   ├── engine.go
│   │   ├── engine_boundaries_test.go
│   │   ├── engine_test.go
│   │   └── model.go
│   └── README.md
├── LICENSE
├── Makefile
├── opencode.jsonc
├── README.md
├── skills
│   └── ste-writing
│       └── SKILL.md
└── specs
    └── 001-three-finger-drag
        ├── plan.md
        ├── results-summary.md
        ├── spec.md
        └── tasks.md

9 directories, 36 files
```

## License

This project uses the MIT License. See the LICENSE file for details.
