#ifndef INPUTPROBE_NATIVE_H
#define INPUTPROBE_NATIVE_H

#include <libevdev/libevdev.h>
#include <libevdev/libevdev-uinput.h>

struct probe_device {
    struct libevdev *dev;
    int fd;
    int grabbed;
};

struct probe_virtual {
    struct libevdev_uinput *dev;
    int fd;
};

struct probe_summary {
    int contacts;
    int palms;
    int buttons;
};

int probe_open(const char *path, struct probe_device **out);
void probe_close(struct probe_device *device);
int probe_grab(struct probe_device *device);
int probe_next(struct probe_device *device, struct input_event *event);
int probe_summary(struct probe_device *device, struct probe_summary *out);
int probe_create(struct probe_device *source, int touchpad, const char *name,
                 struct probe_virtual **out);
void probe_destroy(struct probe_virtual *device);
int probe_write(struct probe_virtual *device, unsigned int type,
                unsigned int code, int value);
const char *probe_node(struct probe_virtual *device);
int probe_is_touchpad(struct probe_device *device);

#endif
