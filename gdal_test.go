// Copyright 2011 go-gdal. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gdal

import (
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTiffDriver(t *testing.T) {
	_, err := GetDriverByName("GTiff")
	if err != nil {
		t.Error(err)
	}
}

func TestMissingMetadata(t *testing.T) {
	ds, err := Open("testdata/tiles.gpkg", ReadOnly)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}

	metadata := ds.Metadata("something-that-wont-exist")
	if len(metadata) != 0 {
		t.Errorf("got %d items, want 0", len(metadata))
	}
}

func TestRegenerateOverviews(t *testing.T) {
	memDrv, err := GetDriverByName("MEM")
	if err != nil {
		t.Error(err)
	}
	// 4x4 pixel size
	ds, err := Open("testdata/smallgeo.tif", ReadOnly)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	redband := ds.RasterBand(1)

	// create 1x1 pixelled tiff which would store overview of  redband
	dstile := memDrv.Create("", 1, 1, 3, Byte, nil)
	redbandDsTile := dstile.RasterBand(1)
	greenbandDsTile := dstile.RasterBand(2)
	blueBandDsTile := dstile.RasterBand(3)
	bands := []RasterBand{
		redbandDsTile,
		greenbandDsTile,
		blueBandDsTile,
	}
	err = redband.RegenerateOverviews(3, &bands[0], "average", DummyProgress, nil)
	if err != nil {
		t.Error(err)
	}
	// generate PNG
	outDrv, err := GetDriverByName("PNG")
	if err != nil {
		panic(err)
	}
	outDrv.CreateCopy("testdata/temp.png", dstile, 0, nil, nil, nil)

	imgFile, err := os.Open("testdata/temp.png")
	if err != nil {
		t.Error(err)
	}
	imgData, err := png.Decode(imgFile)
	if err != nil {
		t.Error(err)
	}

	// assert that png is of 1x1 pixel
	assert.Equal(t, imgData.Bounds().Min.X, 0)
	assert.Equal(t, imgData.Bounds().Min.Y, 0)
	assert.Equal(t, imgData.Bounds().Max.X, 1)
	assert.Equal(t, imgData.Bounds().Max.Y, 1)

	// assert that image contains values RGB(32896, 32896, 32896)
	// 32896 is the computed average of r, g, b values from smallgeo.tif
	r, g, b, _ := imgData.At(0, 0).RGBA()
	assert.Equal(t, int(r), 32896)
	assert.Equal(t, int(g), 32896)
	assert.Equal(t, int(b), 32896)

	// remove temp.png
	if err := os.Remove("testdata/temp.png"); err != nil {
		t.Error(err)
	}
}

func TestVSIReadDirRecursive(t *testing.T) {
	// GDAL returns NULL for an empty directory: this used to dereference it.
	assert.Nil(t, VSIReadDirRecursive("/vsimem/osr_test_empty/"))

	driver, err := GetDriverByName("GTiff")
	require.NoError(t, err)
	ds := driver.Create("/vsimem/osr_test/a.tif", 1, 1, 1, Byte, nil)
	ds.Close()
	defer VSIGetMemFileBuffer("/vsimem/osr_test/a.tif", true)
	assert.Equal(t, []string{"a.tif"}, VSIReadDirRecursive("/vsimem/osr_test/"))
}
