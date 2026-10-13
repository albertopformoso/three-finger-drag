#include "native.h"

#include <errno.h>
#include <fcntl.h>
#include <poll.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

int probe_open(const char *path, struct probe_device **out) {
    struct probe_device *p = calloc(1, sizeof(*p));
    if (!p) return -ENOMEM;
    p->fd = open(path, O_RDONLY | O_NONBLOCK | O_CLOEXEC);
    if (p->fd < 0) {
        int rc = -errno;
        free(p);
        return rc;
    }
    int rc = libevdev_new_from_fd(p->fd, &p->dev);
    if (rc == 0) rc = libevdev_set_clock_id(p->dev, CLOCK_MONOTONIC);
    if (rc < 0) {
        probe_close(p);
        return rc;
    }
    *out = p;
    return 0;
}

void probe_close(struct probe_device *p) {
    if (!p) return;
    if (p->grabbed) libevdev_grab(p->dev, LIBEVDEV_UNGRAB);
    if (p->dev) libevdev_free(p->dev);
    if (p->fd >= 0) close(p->fd);
    free(p);
}

int probe_is_touchpad(struct probe_device *p) {
    return libevdev_has_property(p->dev, INPUT_PROP_POINTER) &&
        libevdev_has_event_code(p->dev, EV_ABS, ABS_MT_POSITION_X) &&
        libevdev_has_event_code(p->dev, EV_ABS, ABS_MT_POSITION_Y) &&
        libevdev_has_event_code(p->dev, EV_ABS, ABS_MT_TRACKING_ID) &&
        libevdev_get_num_slots(p->dev) >= 3 &&
        strncmp(libevdev_get_name(p->dev), "tfd-probe-", 10) != 0;
}

int probe_grab(struct probe_device *p) {
    int rc = libevdev_grab(p->dev, LIBEVDEV_GRAB);
    if (rc == 0) p->grabbed = 1;
    return rc;
}

int probe_next(struct probe_device *p, struct input_event *event) {
    int rc = libevdev_next_event(p->dev, LIBEVDEV_READ_FLAG_NORMAL, event);
    if (rc != -EAGAIN) return rc;
    struct pollfd fds = {.fd = p->fd, .events = POLLIN};
    rc = poll(&fds, 1, 20);
    if (rc < 0) return errno == EINTR ? -EAGAIN : -errno;
    if (fds.revents & (POLLHUP | POLLERR | POLLNVAL)) return -ENODEV;
    if (rc == 0) return -EAGAIN;
    return libevdev_next_event(p->dev, LIBEVDEV_READ_FLAG_NORMAL, event);
}

int probe_summary(struct probe_device *p, struct probe_summary *out) {
    memset(out, 0, sizeof(*out));
    int slots = libevdev_get_num_slots(p->dev);
    for (int i = 0; i < slots; i++) {
        if (libevdev_get_slot_value(p->dev, i, ABS_MT_TRACKING_ID) < 0) continue;
        out->contacts++;
        if (libevdev_has_event_code(p->dev, EV_ABS, ABS_MT_TOOL_TYPE) &&
            libevdev_get_slot_value(p->dev, i, ABS_MT_TOOL_TYPE) == MT_TOOL_PALM)
            out->palms++;
    }
    for (int code = BTN_LEFT; code <= BTN_TASK; code++) {
        if (libevdev_has_event_code(p->dev, EV_KEY, code) &&
            libevdev_get_event_value(p->dev, EV_KEY, code)) out->buttons++;
    }
    return 0;
}

static int axis(struct libevdev *dev, int code, int max, int res, int value) {
    struct input_absinfo info = {.minimum = 0, .maximum = max,
        .resolution = res, .value = value};
    return libevdev_enable_event_code(dev, EV_ABS, code, &info);
}

static int synthetic(struct libevdev *dev, int touchpad) {
    int rc = libevdev_enable_event_code(dev, EV_KEY, BTN_LEFT, NULL);
    if (rc < 0) return rc;
    if (!touchpad) {
        rc = libevdev_enable_event_code(dev, EV_KEY, BTN_RIGHT, NULL);
        if (rc < 0) return rc;
        rc = libevdev_enable_event_code(dev, EV_REL, REL_X, NULL);
        if (rc < 0) return rc;
        return libevdev_enable_event_code(dev, EV_REL, REL_Y, NULL);
    }
    rc = libevdev_enable_property(dev, INPUT_PROP_POINTER);
    if (rc < 0) return rc;
    rc = libevdev_enable_property(dev, INPUT_PROP_BUTTONPAD);
    if (rc < 0) return rc;
    int keys[] = {BTN_TOUCH, BTN_TOOL_FINGER, BTN_TOOL_DOUBLETAP,
        BTN_TOOL_TRIPLETAP, BTN_TOOL_QUADTAP, BTN_TOOL_QUINTTAP};
    for (unsigned int i = 0; i < sizeof(keys)/sizeof(keys[0]); i++) {
        rc = libevdev_enable_event_code(dev, EV_KEY, keys[i], NULL);
        if (rc < 0) return rc;
    }
    int codes[] = {ABS_X, ABS_Y, ABS_MT_SLOT, ABS_MT_POSITION_X,
        ABS_MT_POSITION_Y, ABS_MT_TOOL_TYPE, ABS_MT_TRACKING_ID};
    int maxima[] = {1639, 739, 4, 1639, 739, 2, 65535};
    int resolutions[] = {12, 12, 0, 12, 12, 0, 0};
    for (unsigned int i = 0; i < sizeof(codes)/sizeof(codes[0]); i++) {
        rc = axis(dev, codes[i], maxima[i], resolutions[i], 0);
        if (rc < 0) return rc;
    }
    return 0;
}

static int copy_caps(struct libevdev *from, struct libevdev *to) {
    for (int prop = 0; prop <= INPUT_PROP_MAX; prop++) {
        if (libevdev_has_property(from, prop)) {
            int rc = libevdev_enable_property(to, prop);
            if (rc < 0) return rc;
        }
    }
    for (int type = 1; type <= EV_MAX; type++) {
        if (!libevdev_has_event_type(from, type)) continue;
        for (int code = 0; code <= libevdev_event_type_get_max(type); code++) {
            if (!libevdev_has_event_code(from, type, code)) continue;
            // Current physical coordinates are not copied into the virtual device.
            struct input_absinfo info;
            const void *data = NULL;
            if (type == EV_ABS) {
                info = *libevdev_get_abs_info(from, code);
                info.value = 0;
                data = &info;
            }
            int rc = libevdev_enable_event_code(to, type, code, data);
            if (rc < 0) return rc;
        }
    }
    return 0;
}

int probe_create(struct probe_device *source, int touchpad, const char *name,
                 struct probe_virtual **out) {
    struct libevdev *description = libevdev_new();
    if (!description) return -ENOMEM;
    libevdev_set_name(description, name);
    libevdev_set_phys(description, "tfd-probe/virtual");
    libevdev_set_id_bustype(description, BUS_VIRTUAL);
    libevdev_set_id_vendor(description, 0x1234);
    libevdev_set_id_product(description, touchpad ? 0x1001 : 0x1002);
    int rc = source ? copy_caps(source->dev, description) : synthetic(description, touchpad);
    struct probe_virtual *p = NULL;
    if (rc == 0) {
        p = calloc(1, sizeof(*p));
        if (!p) rc = -ENOMEM;
        else p->fd = -1;
    }
    if (rc == 0) {
        p->fd = open("/dev/uinput", O_RDWR | O_NONBLOCK | O_CLOEXEC);
        if (p->fd < 0) rc = -errno;
    }
    if (rc == 0) rc = libevdev_uinput_create_from_device(description, p->fd, &p->dev);
    libevdev_free(description);
    if (rc < 0) {
        if (p && p->fd >= 0) close(p->fd);
        free(p);
        return rc;
    }
    *out = p;
    return 0;
}

void probe_destroy(struct probe_virtual *p) {
    if (!p) return;
    libevdev_uinput_destroy(p->dev);
    close(p->fd);
    free(p);
}

int probe_write(struct probe_virtual *p, unsigned int type, unsigned int code, int value) {
    return libevdev_uinput_write_event(p->dev, type, code, value);
}

const char *probe_node(struct probe_virtual *p) {
    return libevdev_uinput_get_devnode(p->dev);
}
