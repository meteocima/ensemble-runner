#!/bin/bash
DIR_ROTATE_DAYS=2
TARBALL_DELETION_DAYS=30
DIR=`realpath $1`

cd $DIR
echo "compressing $DIR dirs that are $DIR_ROTATE_DAYS days old...";
for SUBDIR in $(find ./ -maxdepth 1 -mindepth 1 -type d -mtime +"$((DIR_ROTATE_DAYS - 1))" | sort); do
  if [[ $SUBDIR == "./dewetra" ]]; then
    continue
  fi

  if [[ $SUBDIR == "./regrids" ]]; then
    continue
  fi

  REALDIR=`realpath $DIR/$SUBDIR`
  echo -n "compressing $REALDIR ... ";
  if tar czf "$SUBDIR.tar.gz" "$SUBDIR"; then
    echo "done" && rm -rf "$SUBDIR";
  else
    echo "failed";
  fi
done

echo "removing $DIR .tar.gz files that are $TARBALL_DELETION_DAYS days old..."
for FILE in $(find ./ -maxdepth 1 -type f -mtime +"$((TARBALL_DELETION_DAYS - 1))" -name "*.tar.gz" | sort); do
  echo -n "removing $DIR/$FILE ... ";
  if rm -f "$DIR/$FILE"; then
    echo "done";
  else
    echo "failed";
  fi
done
