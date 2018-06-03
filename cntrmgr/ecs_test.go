package cntrmgr

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_StrArrToStrPointerArr(t *testing.T) {
	//func strArrToStrPointerArr(strArr []string) []*strin

	t.Run("EmptyArray", func(t *testing.T) {
		strArr := []string{}
		strPointerArr := strArrToStrPointerArr(strArr)

		assert.Equal(t, []*string {}, strPointerArr)
	})

	t.Run("SingleArray", func(t *testing.T) {
		a := "a"
		strArr := []string{a}
		strPointerArr := strArrToStrPointerArr(strArr)

		assert.Equal(t, []*string {&a}, strPointerArr)
	})

	t.Run("DoubleArray", func(t *testing.T) {
		a := "a"
		b := "b"
		strArr := []string{a, b}
		strPointerArr := strArrToStrPointerArr(strArr)

		assert.Equal(t, []*string {&a, &b}, strPointerArr)
	})

	t.Run("MultipleArray", func(t *testing.T) {
		a := "a"
		b := "b"
		c := "c"
		d := "d"
		strArr := []string{a, b, d, c}
		strPointerArr := strArrToStrPointerArr(strArr)

		assert.Equal(t, []*string {&a, &b, &d, &c}, strPointerArr)
	})

}
