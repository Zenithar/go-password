#! /bin/sh
set -e

case ${1} in
  app:server)

    case ${1} in
      app:server)
        shift 1
        exec /sbin/tini -- /sbin/su-exec nobody /usr/bin/password_server $@
        ;;
    esac
    ;;

  app:help)
    echo "Available options:"
    echo " app:server       - Starts the server"
    echo " app:help         - Displays the help"
    echo " [command]        - Execute the specified command, eg. bash."
    ;;
  *)
    exec "$@"
    ;;
esac
