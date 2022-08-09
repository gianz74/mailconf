#!/bin/sh

if [ ! -d "/tmp/mail" ] ; then
    mkdir -p "/tmp/mail"
fi

if mu index --lazy-check
then test -f /tmp/mail/mu_reindex_now && rm /tmp/mail/mu_reindex_now
else touch /tmp/mail/mu_reindex_now
fi

exit 0
