#!/bin/sh
CONF="/etc/config/qpkg.conf"
QPKG_NAME="BeszelAgent"
QPKG_ROOT=$(/sbin/getcfg "$QPKG_NAME" Install_Path -f "$CONF")
CGI_LINK="/home/httpd/cgi-bin/qpkg/${QPKG_NAME}"
CTL_BIN="${QPKG_ROOT}/www/api.cgi"

export QNAP_QPKG="$QPKG_NAME"
export BESZEL_QPKG_ROOT="$QPKG_ROOT"

ensure_web_link() {
    /bin/mkdir -p /home/httpd/cgi-bin/qpkg
    /bin/rm -rf "$CGI_LINK"
    /bin/ln -sf "${QPKG_ROOT}/www" "$CGI_LINK"
}

ensure_smartctl_link() {
    SMARTCTL_SRC=""
    if [ -x "${QPKG_ROOT}/bin/smartctl" ]; then
        SMARTCTL_SRC="${QPKG_ROOT}/bin/smartctl"
    elif [ -x "${QPKG_ROOT}/data/bin/smartctl" ]; then
        SMARTCTL_SRC="${QPKG_ROOT}/data/bin/smartctl"
    fi
    if [ -n "$SMARTCTL_SRC" ] && [ ! -e /usr/local/sbin/smartctl ]; then
        /bin/mkdir -p /usr/local/sbin
        /bin/ln -sf "$SMARTCTL_SRC" /usr/local/sbin/smartctl
    fi
}

remove_web_link() {
    /bin/rm -rf "$CGI_LINK"
}

remove_smartctl_link() {
    if [ -L /usr/local/sbin/smartctl ]; then
        TARGET=$(readlink /usr/local/sbin/smartctl 2>/dev/null || true)
        case "$TARGET" in
          "${QPKG_ROOT}"/*)
            /bin/rm -f /usr/local/sbin/smartctl
            ;;
        esac
    fi
}

case "$1" in
  start)
    ENABLED=$(/sbin/getcfg "$QPKG_NAME" Enable -u -d FALSE -f "$CONF")
    if [ "$ENABLED" != "TRUE" ]; then
        echo "$QPKG_NAME is disabled."
        exit 1
    fi
    /bin/mkdir -p "${QPKG_ROOT}/data/bin" "${QPKG_ROOT}/data/agent-data" "${QPKG_ROOT}/data/logs" "${QPKG_ROOT}/data/run"
    /bin/chmod 755 "$CTL_BIN" 2>/dev/null || true
    if [ -f "${QPKG_ROOT}/bin/smartctl" ]; then
        /bin/chmod 755 "${QPKG_ROOT}/bin/smartctl" 2>/dev/null || true
    fi
    ensure_web_link
    ensure_smartctl_link
    "$CTL_BIN" ctl start-if-configured || true
    ;;

  stop)
    if [ -x "$CTL_BIN" ]; then
        "$CTL_BIN" ctl stop || true
    fi
    remove_web_link
    ;;

  restart)
    "$0" stop
    "$0" start
    ;;

  status)
    if [ -x "$CTL_BIN" ]; then
        "$CTL_BIN" ctl status
        exit $?
    fi
    exit 1
    ;;

  remove)
    if [ -x "$CTL_BIN" ]; then
        "$CTL_BIN" ctl stop || true
    fi
    remove_web_link
    remove_smartctl_link
    ;;

  *)
    echo "Usage: $0 {start|stop|restart|status|remove}"
    exit 1
esac

exit 0
