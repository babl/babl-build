#!/bin/bash -x


if [[ $QUALITY && ${QUALITY-x} ]]; then
  quality="-quality $QUALITY"
fi

if [[ $DPI && ${DPI-x} ]]; then
  dpi="-density $DPI"
fi

format=${FORMAT-png}

convert -units PixelsPerInch -background none - -resize "${WIDTH}x${HEIGHT}^" -gravity center -extent "${WIDTH}x${HEIGHT}" $dpi $quality $format:-
