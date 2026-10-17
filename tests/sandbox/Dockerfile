FROM alpine:3.24
RUN apk add --no-cache foot fuzzel wtype waybar swaync sway swaybg grim dbus ttf-dejavu libnotify mesa-dri-gallium python3 py3-pillow tesseract-ocr tesseract-ocr-data-eng \
    && adduser -D tester \
    && mkdir -p /tmp/runtime /tmp/results \
    && chown tester:tester /tmp/runtime /tmp/results \
    && chmod 700 /tmp/runtime \
    && cp /usr/bin/sway /usr/local/bin/sway
# The copy strips Alpine's file capabilities; no elevated capabilities are needed.
COPY scenario.py visual.py desktop.py /opt/sandbox/
COPY sway.conf /opt/sandbox/sway.conf
USER tester
ENV HOME=/home/tester XDG_RUNTIME_DIR=/tmp/runtime \
    WLR_BACKENDS=headless WLR_RENDERER=pixman WLR_LIBINPUT_NO_DEVICES=1 \
    GDK_BACKEND=wayland GSK_RENDERER=cairo
ENTRYPOINT ["dbus-run-session", "--", "python3", "/opt/sandbox/scenario.py"]
