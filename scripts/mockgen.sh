#!/bin/bash

# should be in your path
MOCKGEN=mockgen
SED=sed
GOFMT=gofmt
MKDIR=mkdir

generate_mock() {
    SRC=$1
    PKG=$(dirname $SRC)
    DST=$PKG/mock_$(basename $SRC)

    $MKDIR -p $(dirname $DST)
    $MOCKGEN -source ./$SRC -destination ./$DST -package $(basename $PKG)
    $GOFMT -w ./$DST
}

generate_vendor_mock() {
    PKG=$1
    INTERFACES=$2
    DST=mocks/$PKG/mock_$(basename $PKG).go

    $MKDIR -p $(dirname $DST)
    $MOCKGEN -destination ./$DST -package $(basename $PKG) $PKG $INTERFACES
    $GOFMT -w ./$DST
}

# generate vendor mocks
generate_vendor_mock net/http Handler