package gdal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The exports return strings GDAL allocates, which the wrappers now free with CPLFree after
// copying. These check the copies are intact; the freeing itself is what removes the leak.
func TestSpatialReferenceExports(t *testing.T) {
	sr := CreateSpatialReference("")
	defer sr.Destroy()
	require.NoError(t, sr.FromEPSG(3857))

	wkt, err := sr.ToWKT()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(wkt, `PROJCS["WGS 84 / Pseudo-Mercator"`), wkt)

	pretty, err := sr.ToPrettyWKT(false)
	require.NoError(t, err)
	assert.Contains(t, pretty, "\n")
	assert.Contains(t, pretty, "Pseudo-Mercator")

	proj4, err := sr.ToProj4()
	require.NoError(t, err)
	assert.Contains(t, proj4, "+proj=merc")

	// GDAL's XML export has no Pseudo-Mercator, so XML and MapInfo use EPSG:4326.
	sr4326 := CreateSpatialReference("")
	defer sr4326.Destroy()
	require.NoError(t, sr4326.FromEPSG(4326))

	xml, err := sr4326.ToXML()
	require.NoError(t, err)
	assert.Contains(t, xml, "<gml:GeographicCRS")

	mi, err := sr4326.ToMICoordSys()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(mi, "Earth Projection"), mi)
}

func TestSpatialReferenceFromEPSGUnknown(t *testing.T) {
	sr := CreateSpatialReference("")
	defer sr.Destroy()
	assert.Error(t, sr.FromEPSG(999999))
}

func TestGeometryExports(t *testing.T) {
	geom, err := CreateFromWKT("POINT (1 2)", SpatialReference{})
	require.NoError(t, err)
	defer geom.Destroy()

	assert.Equal(t, `{ "type": "Point", "coordinates": [ 1.0, 2.0 ] }`, geom.ToJSON())
	assert.Equal(t, `{ "type": "Point", "coordinates": [ 1.0, 2.0 ] }`, geom.ToJSON_ex(nil))
	assert.Equal(t, `<gml:Point><gml:coordinates>1,2</gml:coordinates></gml:Point>`, geom.ToGML())
	assert.Equal(t, `<gml:Point><gml:coordinates>1,2</gml:coordinates></gml:Point>`, geom.ToGML_Ex(nil))
}

func TestDataTypeSize(t *testing.T) {
	for dt, bits := range map[DataType]int{
		Byte:     8,
		Int16:    16,
		Float32:  32,
		Float64:  64,
		CFloat64: 128,
	} {
		assert.Equal(t, bits, dt.Size(), dt.Name())
	}
}
