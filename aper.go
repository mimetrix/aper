package aper

import (
    //"github.com/davecgh/go-spew/spew"
    "encoding/json"
	"fmt"
    "strings"
    //"encoding/hex"
	"path"
	"reflect"
	"runtime"
	"github.com/free5gc/aper/logger"
)

type PerBitData struct {
	bytes      []byte
	byteOffset uint64
	bitsOffset uint
}

type AperDecoder interface {
    AperDecode(pd *PerBitData, params FieldParameters ) error
}

func perTrace(level int, s string) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		logger.AperLog.Debugln(s)
	} else {
		logger.AperLog.Debugf("%s (%s:%d)\n", s, path.Base(file), line)
	}
}

func perBitLog(numBits uint64, byteOffset uint64, bitsOffset uint, value interface{}) string {
	if reflect.TypeOf(value).Kind() == reflect.Uint64 {
		return fmt.Sprintf("  [PER got %2d bits, byteOffset(after): %d, bitsOffset(after): %d, value: 0x%0x]",
			numBits, byteOffset, bitsOffset, reflect.ValueOf(value).Uint())
	}
	return fmt.Sprintf("  [PER got %2d bits, byteOffset(after): %d, bitsOffset(after): %d, value: 0x%0x]",
		numBits, byteOffset, bitsOffset, reflect.ValueOf(value).Bytes())
}

// GetBitString is to get BitString with desire size from source byte array with bit offset
func GetBitString(srcBytes []byte, bitsOffset uint, numBits uint) (dstBytes []byte, err error) {
	bitsLeft := uint(len(srcBytes))*8 - bitsOffset

	if numBits > bitsLeft {
		err = fmt.Errorf("Get bits overflow, requireBits: %d, leftBits: %d", numBits, bitsLeft)
		return
	}

	byteLen := (bitsOffset + numBits + 7) >> 3
	numBitsByteLen := (numBits + 7) >> 3
	dstBytes = make([]byte, numBitsByteLen)
	numBitsMask := byte(0xff)
	if modEight := numBits & 0x7; modEight != 0 {
		numBitsMask <<= uint8(8 - (modEight))
	}
	for i := 1; i < int(byteLen); i++ {
		dstBytes[i-1] = srcBytes[i-1]<<bitsOffset | srcBytes[i]>>(8-bitsOffset)
	}
	if byteLen == numBitsByteLen {
		dstBytes[byteLen-1] = srcBytes[byteLen-1] << bitsOffset
	}
	dstBytes[numBitsByteLen-1] &= numBitsMask
	return
}

// GetFewBits is to get Value with desire few bits from source byte with bit offset
// func GetFewBits(srcByte byte, bitsOffset uint, numBits uint) (value uint64, err error) {

// 	if numBits == 0 {
// 		value = 0
// 		return
// 	}
// 	bitsLeft := 8 - bitsOffset
// 	if bitsLeft < numBits {
// 		err = fmt.Errorf("Get bits overflow, requireBits: %d, leftBits: %d", numBits, bitsLeft)
// 		return
// 	}
// 	if bitsOffset == 0 {
// 		value = uint64(srcByte >> (8 - numBits))
// 	} else {
// 		value = uint64((srcByte << bitsOffset) >> (8 - numBits))
// 	}
// 	return
// }

// GetBitsValue is to get Value with desire bits from source byte array with bit offset
func GetBitsValue(srcBytes []byte, bitsOffset uint, numBits uint) (value uint64, err error) {
	var dstBytes []byte
	dstBytes, err = GetBitString(srcBytes, bitsOffset, numBits)
	if err != nil {
		return
	}
	for i, j := 0, numBits; j >= 8; i, j = i+1, j-8 {
		value <<= 8
		value |= uint64(uint(dstBytes[i]))
	}
	if numBitsOff := (numBits & 0x7); numBitsOff != 0 {
		var mask uint = (1 << numBitsOff) - 1
		value <<= numBitsOff
		value |= uint64(uint(dstBytes[len(dstBytes)-1]>>(8-numBitsOff)) & mask)
	}
	return
}

func (pd *PerBitData) bitCarry() {
	pd.byteOffset += uint64(pd.bitsOffset >> 3)
	pd.bitsOffset = pd.bitsOffset & 0x07
}

func (pd *PerBitData) getBitString(numBits uint) (dstBytes []byte, err error) {
	dstBytes, err = GetBitString(pd.bytes[pd.byteOffset:], pd.bitsOffset, numBits)
	if err != nil {
		return
	}
	pd.bitsOffset += numBits

	pd.bitCarry()
	perTrace(1, perBitLog(uint64(numBits), pd.byteOffset, pd.bitsOffset, dstBytes))
	return
}

func (pd *PerBitData) GetBitsValue(numBits uint) (value uint64, err error) {
	value, err = GetBitsValue(pd.bytes[pd.byteOffset:], pd.bitsOffset, numBits)
	if err != nil {
		return
	}
	pd.bitsOffset += numBits
	pd.bitCarry()
	perTrace(1, perBitLog(uint64(numBits), pd.byteOffset, pd.bitsOffset, value))
	return
}

func (pd *PerBitData) parseAlignBits() error {
	if (pd.bitsOffset & 0x7) > 0 {
		alignBits := 8 - ((pd.bitsOffset) & 0x7)
		perTrace(2, fmt.Sprintf("Aligning %d bits", alignBits))
		if val, err := pd.GetBitsValue(alignBits); err != nil {
			return err
		} else if val != 0 {
			if skipPaddingCheck {
				perTrace(2, "Align Bit is not zero")
				perTrace(1, perBitLog(uint64(alignBits), pd.byteOffset, pd.bitsOffset, val))
			} else {
				return fmt.Errorf("Align Bit is not zero")
			}
		}
	} else if pd.bitsOffset != 0 {
		pd.bitCarry()
	}
	return nil
}

func (pd *PerBitData) parseConstraintValue(valueRange int64) (value uint64, err error) {
	perTrace(3, fmt.Sprintf("Getting Constraint Value with range %d", valueRange))

	var bytes uint
	if valueRange <= 255 {
		if valueRange < 0 {
			err = fmt.Errorf("Value range is negative")
			return
		}
		var i uint
		// 1 ~ 8 bits
		for i = 1; i <= 8; i++ {
			upper := 1 << i
			if int64(upper) >= valueRange {
				break
			}
		}
		value, err = pd.GetBitsValue(i)
		return
	} else if valueRange == 256 {
		bytes = 1
	} else if valueRange <= 65536 {
		bytes = 2
	} else {
		err = fmt.Errorf("Constraint Value is large than 65536")
		return
	}
	if err = pd.parseAlignBits(); err != nil {
		return
	}
	value, err = pd.GetBitsValue(bytes * 8)
	return value, err
}

func (pd *PerBitData) parseSemiConstrainedWholeNumber(lb uint64) (value uint64, err error) {
	var repeat bool
	var length uint64
	if length, err = pd.parseLength(-1, &repeat); err != nil {
		return
	}
	if length > 8 || repeat {
		err = fmt.Errorf("Too long length: %d", length)
		return
	}
	if value, err = pd.GetBitsValue(uint(length) * 8); err != nil {
		return
	}
	value += lb
	return
}

func (pd *PerBitData) parseNormallySmallNonNegativeWholeNumber() (value uint64, err error) {
	var notSmallFlag uint64
	if notSmallFlag, err = pd.GetBitsValue(1); err != nil {
		return
	}
	if notSmallFlag == 1 {
		if value, err = pd.parseSemiConstrainedWholeNumber(0); err != nil {
			return
		}
	} else {
		if value, err = pd.GetBitsValue(6); err != nil {
			return
		}
	}
	return
}

func (pd *PerBitData) parseLength(sizeRange int64, repeat *bool) (value uint64, err error) {
	*repeat = false
	if sizeRange <= 65536 && sizeRange > 0 {
		return pd.parseConstraintValue(sizeRange)
	}

	if err = pd.parseAlignBits(); err != nil {
		return
	}
	firstByte, err := pd.GetBitsValue(8)
	if err != nil {
		return
	}
	if (firstByte & 128) == 0 { // #10.9.3.6
		value = firstByte & 0x7F
		return
	} else if (firstByte & 64) == 0 { // #10.9.3.7
		var secondByte uint64
		if secondByte, err = pd.GetBitsValue(8); err != nil {
			return
		}
		value = ((firstByte & 63) << 8) | secondByte
		return
	}
	firstByte &= 63
	if firstByte < 1 || firstByte > 4 {
		err = fmt.Errorf("Parse Length Out of Constraint")
		return
	}
	*repeat = true
	value = 16384 * firstByte
	return value, err
}

func GetHexString(bytes []byte, sep string) string{
    if len(bytes) > 0 {
        
        hexvals := make([]string,0, len(bytes))
        for idx := range bytes{
            chr := fmt.Sprintf("%.2x",bytes[idx])
            if chr != "" { 
                hexvals = append(hexvals,chr)
            }
        }
        return strings.Join(hexvals,sep)
    }

    return "" 
}


func (pd *PerBitData) Exported() int{
    return 999
}

func (pd *PerBitData) notExported() int{
    return 999
}

func (pd *PerBitData) ParseBitString(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (BitString, error) {

    /*
    fmt.Printf("\n\tcalled parseBitString %v - %v - %v",extensed, *lowerBoundPtr, *upperBoundPtr)
    spew.Dump(pd)
    fmt.Println()
*/
	var lb, ub, sizeRange int64 = 0, -1, -1
	if !extensed {
		if lowerBoundPtr != nil {
			lb = *lowerBoundPtr
		}
		if upperBoundPtr != nil {
			ub = *upperBoundPtr
			sizeRange = ub - lb + 1
		}
	}
	if ub > 65535 {
		sizeRange = -1
	}
	// initailization
	bitString := BitString{[]byte{},"", 0}
	//bitString := BitString{[]byte{}, 0}
	// lowerbound == upperbound
	if sizeRange == 1 {
		sizes := uint64(ub+7) >> 3
		bitString.BitLength = uint64(ub)
		perTrace(2, fmt.Sprintf("Decoding BIT STRING size %d", ub))
		if sizes > 2 {
			if err := pd.parseAlignBits(); err != nil {
				return bitString, err
			}
			if (pd.byteOffset + sizes) > uint64(len(pd.bytes)) {
				err := fmt.Errorf("PER data out of range")
				return bitString, err
			}
			bitString.Bytes = pd.bytes[pd.byteOffset : pd.byteOffset+sizes]
			pd.byteOffset += sizes
			pd.bitsOffset = uint(ub & 0x7)
			if pd.bitsOffset > 0 {
				pd.byteOffset--
			}
			perTrace(1, perBitLog(uint64(ub), pd.byteOffset, pd.bitsOffset, bitString.Bytes))
		} else {
			if bytes, err := pd.getBitString(uint(ub)); err != nil {
				logger.AperLog.Warnf("PD GetBitString error: %+v", err)
				return bitString, err
			} else {
				bitString.Bytes = bytes
			}
		}
		perTrace(2, fmt.Sprintf("Decoded BIT STRING (length = %d): %0.8b", ub, bitString.Bytes))
        bitString.ByteString  = GetHexString(bitString.Bytes,"")
		return bitString, nil
	}
	repeat := false
	for {
		var rawLength uint64
		if length, err := pd.parseLength(sizeRange, &repeat); err != nil {
			return bitString, err
		} else {
			rawLength = length
		}
		rawLength += uint64(lb)
		perTrace(2, fmt.Sprintf("Decoding BIT STRING size %d", rawLength))
		if rawLength == 0 {
			return bitString, nil
		}
		sizes := (rawLength + 7) >> 3
		if err := pd.parseAlignBits(); err != nil {
			return bitString, err
		}

		if (pd.byteOffset + sizes) > uint64(len(pd.bytes)) {
			err := fmt.Errorf("PER data out of range")
			return bitString, err
		}
		bitString.Bytes = append(bitString.Bytes, pd.bytes[pd.byteOffset:pd.byteOffset+sizes]...)
		bitString.BitLength += rawLength
		pd.byteOffset += sizes
		pd.bitsOffset = uint(rawLength & 0x7)
		if pd.bitsOffset != 0 {
			pd.byteOffset--
		}
		perTrace(1, perBitLog(rawLength, pd.byteOffset, pd.bitsOffset, bitString.Bytes))
		perTrace(2, fmt.Sprintf("Decoded BIT STRING (length = %d): %0.8b", rawLength, bitString.Bytes))

		if !repeat {
			// if err = pd.parseAlignBits(); err != nil {
			// 	return
			// }
			break
		}
	}

    //bitString.HexBytes = hex.EncodeToString(bitString.Bytes)
    bitString.ByteString  = GetHexString(bitString.Bytes,"")
	return bitString, nil
}

func (pd *PerBitData) ParseOctetString(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (
	OctetString, error,
) {

	var lb, ub, sizeRange int64 = 0, -1, -1
	if !extensed {
		if lowerBoundPtr != nil {
			lb = *lowerBoundPtr
		}
		if upperBoundPtr != nil {
			ub = *upperBoundPtr
			sizeRange = ub - lb + 1
		}
	}
	if ub > 65535 {
		sizeRange = -1
	}
	// initailization
	//octetString := OctetString("")
	octetString := OctetString{[]byte{},""}
	// lowerbound == upperbound
	if sizeRange == 1 {
		perTrace(2, fmt.Sprintf("Decoding OCTET STRING size %d", ub))
		if ub > 2 {
			unsignedUB := uint64(ub)
			if err := pd.parseAlignBits(); err != nil {
				return octetString, err
			}
			if (int64(pd.byteOffset) + ub) > int64(len(pd.bytes)) {
				err := fmt.Errorf("per data out of range")
				return octetString, err
			}
			//octetString = pd.bytes[pd.byteOffset : pd.byteOffset+unsignedUB]
			octetString.Bytes = pd.bytes[pd.byteOffset : pd.byteOffset+unsignedUB]
			pd.byteOffset += uint64(ub)
			perTrace(1, perBitLog(8*unsignedUB, pd.byteOffset, pd.bitsOffset, octetString.Bytes))
		} else {
			if octet, err := pd.getBitString(uint(ub * 8)); err != nil {
				return octetString, err
			} else {
				octetString.Bytes = octet
			}
		}
		perTrace(2, fmt.Sprintf("Decoded OCTET STRING (length = %d): 0x%0x", ub, octetString.Bytes))
        octetString.OctetString  = GetHexString(octetString.Bytes,":")
		return octetString, nil
	}
	repeat := false
	for {
		var rawLength uint64
		if length, err := pd.parseLength(sizeRange, &repeat); err != nil {
			return octetString, err
		} else {
			rawLength = length
		}
		rawLength += uint64(lb)
		perTrace(2, fmt.Sprintf("Decoding OCTET STRING size %d", rawLength))
		if rawLength == 0 {
			return octetString, nil
		} else if err := pd.parseAlignBits(); err != nil {
			return octetString, err
		}
		if (rawLength + pd.byteOffset) > uint64(len(pd.bytes)) {
			err := fmt.Errorf("per data out of range ")
			return octetString, err
		}
		octetString.Bytes = append(octetString.Bytes, pd.bytes[pd.byteOffset:pd.byteOffset+rawLength]...)
		pd.byteOffset += rawLength
		perTrace(1, perBitLog(8*rawLength, pd.byteOffset, pd.bitsOffset, octetString.Bytes))
		perTrace(2, fmt.Sprintf("Decoded OCTET STRING (length = %d): 0x%0x", rawLength, octetString.Bytes))
		if !repeat {
			// if err = pd.parseAlignBits(); err != nil {
			// 	return
			// }
			break
		}
	}
    octetString.OctetString  = GetHexString(octetString.Bytes,":")
	return octetString, nil
}

func (pd *PerBitData) parseBool() (value bool, err error) {
	perTrace(3, "Decoding BOOLEAN Value")
	bit, err1 := pd.GetBitsValue(1)
	if err1 != nil {
		err = err1
		return
	}
	if bit == 1 {
		value = true
		perTrace(2, "Decoded BOOLEAN Value : ture")
	} else {
		value = false
		perTrace(2, "Decoded BOOLEAN Value : false")
	}
	return
}

func (pd *PerBitData) parseInteger(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (int64, error) {
	var lb, ub, valueRange int64 = 0, -1, 0
	if !extensed {
		if lowerBoundPtr == nil {
			perTrace(3, "Decoding INTEGER with Unconstraint Value")
			valueRange = -1
		} else {
			lb = *lowerBoundPtr
			if upperBoundPtr != nil {
				ub = *upperBoundPtr
				valueRange = ub - lb + 1
				perTrace(3, fmt.Sprintf("Decoding INTEGER with Value Range(%d..%d)", lb, ub))
			} else {
				perTrace(3, fmt.Sprintf("Decoding INTEGER with Semi-Constraint Range(%d..)", lb))
			}
		}
	} else {
		valueRange = -1
		perTrace(3, "Decoding INTEGER with Extensive Value")
	}
	var rawLength uint
	if valueRange == 1 {
		return ub, nil
	} else if valueRange <= 0 {
		// semi-constraint or unconstraint
		if err := pd.parseAlignBits(); err != nil {
			return int64(0), err
		}
		if pd.byteOffset >= uint64(len(pd.bytes)) {
			return int64(0), fmt.Errorf("per data out of range")
		}
		rawLength = uint(pd.bytes[pd.byteOffset])
		pd.byteOffset++
		perTrace(1, perBitLog(8, pd.byteOffset, pd.bitsOffset, uint64(rawLength)))
	} else if valueRange <= 65536 {
		rawValue, err := pd.parseConstraintValue(valueRange)
		if err != nil {
			return int64(0), err
		} else {
			return int64(rawValue) + lb, nil
		}
	} else {
		// valueRange > 65536
		var byteLen uint
		unsignedValueRange := uint64(valueRange - 1)
		for byteLen = 1; byteLen <= 127; byteLen++ {
			unsignedValueRange >>= 8
			if unsignedValueRange == 0 {
				break
			}
		}
		var i, upper uint
		// 1 ~ 8 bits
		for i = 1; i <= 8; i++ {
			upper = 1 << i
			if upper >= byteLen {
				break
			}
		}
		if tempLength, err := pd.GetBitsValue(i); err != nil {
			return int64(0), err
		} else {
			rawLength = uint(tempLength)
		}
		rawLength++
		if err := pd.parseAlignBits(); err != nil {
			return int64(0), err
		}
	}
	perTrace(2, fmt.Sprintf("Decoding INTEGER Length with %d bytes", rawLength))

	if rawValue, err := pd.GetBitsValue(rawLength * 8); err != nil {
		return int64(0), err
	} else if valueRange < 0 {
		signedBitMask := uint64(1 << (rawLength*8 - 1))
		valueMask := signedBitMask - 1
		// negative
		if rawValue&signedBitMask > 0 {
			return int64((^rawValue)&valueMask+1) * -1, nil
		}
		return int64(rawValue) + lb, nil
	} else {
		return int64(rawValue) + lb, nil
	}
}

func (pd *PerBitData) parseEnumerated(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (value uint64,
	err error,
) {
	if lowerBoundPtr == nil || upperBoundPtr == nil {
		err = fmt.Errorf("ENUMERATED value constraint is error")
		return
	}
	lb, ub := *lowerBoundPtr, *upperBoundPtr
	if lb < 0 || lb > ub {
		err = fmt.Errorf("ENUMERATED value constraint is error")
		return
	}

	if extensed {
		perTrace(2, fmt.Sprintf("Decoding ENUMERATED with Extensive Value of Range(%d..)", ub+1))
		if value, err = pd.parseNormallySmallNonNegativeWholeNumber(); err != nil {
			return
		}
		value += uint64(ub) + 1
	} else {
		perTrace(2, fmt.Sprintf("Decoding ENUMERATED with Value Range(%d..%d)", lb, ub))
		valueRange := ub - lb + 1
		if valueRange > 1 {
			value, err = pd.parseConstraintValue(valueRange)
		}
	}
	perTrace(2, fmt.Sprintf("Decoded ENUMERATED Value : %d", value))
	return
}

func (pd *PerBitData) parseSequenceOf(sizeExtensed bool, params FieldParameters, sliceType reflect.Type) (
	reflect.Value, error,
) {
	var sliceContent reflect.Value
	var lb int64 = 0
	var sizeRange int64
	if params.sizeLowerBound != nil && *params.sizeLowerBound < 65536 {
		lb = *params.sizeLowerBound
	}
	if !sizeExtensed && params.sizeUpperBound != nil && *params.sizeUpperBound < 65536 {
		ub := *params.sizeUpperBound
		sizeRange = ub - lb + 1
		perTrace(3, fmt.Sprintf("Decoding Length of \"SEQUENCE OF\"  with Size Range(%d..%d)", lb, ub))
	} else {
		sizeRange = -1
		perTrace(3, fmt.Sprintf("Decoding Length of \"SEQUENCE OF\" with Semi-Constraint Range(%d..)", lb))
	}

	var numElements uint64
	if sizeRange > 1 {
		if numElementsTmp, err := pd.parseConstraintValue(sizeRange); err != nil {
			logger.AperLog.Warnf("Parse Constraint Value failed: %+v", err)
		} else {
			numElements = numElementsTmp
		}
		numElements += uint64(lb)
	} else if sizeRange == 1 {
		numElements += uint64(lb)
	} else {
		if err := pd.parseAlignBits(); err != nil {
			return sliceContent, err
		}
		if pd.byteOffset >= uint64(len(pd.bytes)) {
			err := fmt.Errorf("per data out of range")
			return sliceContent, err
		}
		numElements = uint64(pd.bytes[pd.byteOffset])
		pd.byteOffset++
		perTrace(1, perBitLog(8, pd.byteOffset, pd.bitsOffset, numElements))
	}
	perTrace(2, fmt.Sprintf("Decoding  \"SEQUENCE OF\" struct %s with len(%d)", sliceType.Elem().Name(), numElements))
	params.sizeExtensible = false
	params.sizeUpperBound = nil
	params.sizeLowerBound = nil
	intNumElements := int(numElements)
	sliceContent = reflect.MakeSlice(sliceType, intNumElements, intNumElements)
	for i := 0; i < intNumElements; i++ {
		err := ParseField(sliceContent.Index(i), pd, params)
		if err != nil {
			return sliceContent, err
		}
	}
	return sliceContent, nil
}

func (pd *PerBitData) getChoiceIndex(extensed bool, upperBoundPtr *int64) (present int, err error) {
	if extensed {
		err = fmt.Errorf("Unsupport value of CHOICE type is in Extensed")
	} else if upperBoundPtr == nil {
		err = fmt.Errorf("The upper bound of CHIOCE is missing")
	} else if ub := *upperBoundPtr; ub < 0 {
		err = fmt.Errorf("The upper bound of CHIOCE is negative")
	} else if rawChoice, err1 := pd.parseConstraintValue(ub + 1); err1 != nil {
		err = err1
	} else {
		perTrace(2, fmt.Sprintf("Decoded Present index of CHOICE is %d + 1", rawChoice))
		present = int(rawChoice) + 1
	}
	return
}

func getReferenceFieldValue(v reflect.Value) (value int64, err error) {
	fieldType := v.Type()
	switch v.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		value = v.Int()
	case reflect.Struct:
		if fieldType.Field(0).Name == "Present" {
			present := int(v.Field(0).Int())
			if present == 0 {
				err = fmt.Errorf("ReferenceField Value present is 0(present's field number)")
			} else if present >= fieldType.NumField() {
				err = fmt.Errorf("Present is bigger than number of struct field")
			} else {
				value, err = getReferenceFieldValue(v.Field(present))
			}
		} else {
			value, err = getReferenceFieldValue(v.Field(0))
		}
	default:
		err = fmt.Errorf("OpenType reference only support INTEGER")
	}
	return
}

func (pd *PerBitData) parseOpenType(v reflect.Value, params FieldParameters) error {
	pdOpenType := &PerBitData{[]byte(""), 0, 0}
	repeat := false
	for {
		var rawLength uint64
		if rawLengthTmp, err := pd.parseLength(-1, &repeat); err != nil {
			return err
		} else {
			rawLength = rawLengthTmp
		}
		if rawLength == 0 {
			break
		} else if err := pd.parseAlignBits(); err != nil {
			return err
		}
		if (rawLength + pd.byteOffset) > uint64(len(pd.bytes)) {
			return fmt.Errorf("per data out of range ")
		}
		pdOpenType.bytes = append(pdOpenType.bytes, pd.bytes[pd.byteOffset:pd.byteOffset+rawLength]...)
		pd.byteOffset += rawLength

		if !repeat {
			if err := pd.parseAlignBits(); err != nil {
				return err
			}
			break
		}
	}
	perTrace(2, fmt.Sprintf("Decoding OpenType %s with (len = %d byte)", v.Type().String(), len(pdOpenType.bytes)))
	err := ParseField(v, pdOpenType, params)
	perTrace(2, fmt.Sprintf("Decoded OpenType %s", v.Type().String()))
	return err
}

// parseField is the main parsing function. Given a byte slice and an offset
// into the array, it will try to parse a suitable ASN.1 value out and store it
// in the given Value. TODO : ObjectIdenfier, handle extension Field
func ParseField(v reflect.Value, pd *PerBitData, params FieldParameters) error {
	fieldType := v.Type()
    //spew.Dump(v.Type().Name())
	// If we have run out of data return error.
	if pd.byteOffset == uint64(len(pd.bytes)) {
		return fmt.Errorf("sequence truncated")
	}
	if v.Kind() == reflect.Ptr {
		ptr := reflect.New(fieldType.Elem())
		v.Set(ptr)
		return ParseField(v.Elem(), pd, params)
	}
	sizeExtensible := false
	valueExtensible := false
	if params.sizeExtensible {
		if bitsValue, err1 := pd.GetBitsValue(1); err1 != nil {
			return err1
		} else if bitsValue != 0 {
			sizeExtensible = true
		}
		perTrace(2, fmt.Sprintf("Decoded Size Extensive Bit : %t", sizeExtensible))
	}
    //spew.Dump(params)
	if params.valueExtensible && v.Kind() != reflect.Slice {
		if bitsValue, err1 := pd.GetBitsValue(1); err1 != nil {
			return err1
		} else if bitsValue != 0 {
			valueExtensible = true
		}
		perTrace(2, fmt.Sprintf("Decoded Value Extensive Bit : %t", valueExtensible))
	}

	// We deal with the structures defined in this package first.
	switch fieldType {
	case BitStringType:
        //fmt.Printf("--pbs--%v\n",sizeExtensible)
        //spew.Dump(pd)
        //fmt.Println("----")
		bitString, err1 := pd.ParseBitString(sizeExtensible, params.sizeLowerBound, params.sizeUpperBound)

		if err1 != nil {
			return err1
		}
		v.Set(reflect.ValueOf(bitString))
		return nil
	case ObjectIdentifierType:
		return fmt.Errorf("Unsupport ObjectIdenfier type")
	case OctetStringType:
        /*
        if params.sizeLowerBound != nil && params.sizeUpperBound != nil {
            fmt.Printf("\nparseoctet:%v - %v - %v",sizeExtensible, *params.sizeLowerBound, *params.sizeUpperBound)
		}
        */
        if octetString, err := pd.ParseOctetString(sizeExtensible, params.sizeLowerBound, params.sizeUpperBound); err != nil {
			return err
		} else {
			v.Set(reflect.ValueOf(octetString))
			return nil
		}
	case EnumeratedType:
		if parsedEnum, err := pd.parseEnumerated(valueExtensible, params.valueLowerBound,
			params.valueUpperBound); err != nil {
			return err
		} else {
			v.SetUint(parsedEnum)
			return nil
		}
	}


	switch val := v; val.Kind() {
	case reflect.Bool:
		if parsedBool, err := pd.parseBool(); err != nil {
			return err
		} else {
			val.SetBool(parsedBool)
			return nil
		}
	case reflect.Int, reflect.Int32, reflect.Int64:
		if parsedInt, err := pd.parseInteger(valueExtensible, params.valueLowerBound, params.valueUpperBound); err != nil {
			return err
		} else {
			val.SetInt(parsedInt)
			perTrace(2, fmt.Sprintf("Decoded INTEGER Value: %d", parsedInt))
			return nil
		}


	case reflect.Struct:

		structType := fieldType
		var structParams []FieldParameters
		var optionalCount uint
		var optionalPresents uint64

		// pass tag for optional
		for i := 0; i < structType.NumField(); i++ {

			if structType.Field(i).PkgPath != "" {
				return fmt.Errorf("struct contains unexported fields : " + structType.Field(i).PkgPath)
			}
			tempParams := parseFieldParameters(structType.Field(i).Tag.Get("aper"))
			// for optional flag
			if tempParams.optional {
				optionalCount++
			}
			structParams = append(structParams, tempParams)
		}

		if optionalCount > 0 {
			if optionalPresentsTmp, err := pd.GetBitsValue(optionalCount); err != nil {
				return err
			} else {
				optionalPresents = optionalPresentsTmp
			}
			perTrace(2, fmt.Sprintf("optionalPresents is %0b", optionalPresents))
		}

		// CHOICE or OpenType
		if structType.NumField() > 0 && structType.Field(0).Name == "Present" {
			var present int = 0
			if params.openType {
				if params.referenceFieldValue == nil {
					return fmt.Errorf("OpenType reference value is empty")
				}
				refValue := *params.referenceFieldValue

				for j, param := range structParams {
					if j == 0 {
						continue
					}
					if param.referenceFieldValue != nil && *param.referenceFieldValue == refValue {
						present = j
						break
					}
				}
				if present == 0 {
					return fmt.Errorf("OpenType reference value does not match any field")
				} else if present >= structType.NumField() {
					return fmt.Errorf("OpenType Present is bigger than number of struct field")
				} else {
					val.Field(0).SetInt(int64(present))
					perTrace(2, fmt.Sprintf("Decoded Present index of OpenType is %d ", present))
					return pd.parseOpenType(val.Field(present), structParams[present])
				}
			} else {
				if presentTmp, err := pd.getChoiceIndex(valueExtensible, params.valueUpperBound); err != nil {
					logger.AperLog.Errorf("pd.getChoiceIndex Error")
				} else {
					present = presentTmp
				}
				val.Field(0).SetInt(int64(present))
				if present == 0 {
					return fmt.Errorf("CHOICE present is 0(present's field number)")
				} else if present >= structType.NumField() {
					return fmt.Errorf("CHOICE Present is bigger than number of struct field")
				} else {
                    if _, ok := val.Field(present).Interface().(AperDecoder); ok {
                        decoderType := reflect.New(val.Field(present).Type().Elem())
                        val.Field(present).Set(decoderType)
                        adInterface := val.Field(present).Interface().(AperDecoder)
                        err := adInterface.AperDecode(pd, structParams[present] )
                        return err
                    } else{ 
					    return ParseField(val.Field(present), pd, structParams[present])
				    }
					//return ParseField(val.Field(present), pd, structParams[present])
                    
                }
			}
		}
        
          
        //spew.Dump(val.Interface())
		for i := 0; i < structType.NumField(); i++ {
            //fmt.Printf("\tField:%d/%d",i,structType.NumField()) 
            if structParams[i].optional && optionalCount > 0 {
				optionalCount--
				if optionalPresents&(1<<optionalCount) == 0 {
					perTrace(3, fmt.Sprintf("Field \"%s\" in %s is OPTIONAL and not present", structType.Field(i).Name, structType))
					continue
				} else {
					perTrace(3, fmt.Sprintf("Field \"%s\" in %s is OPTIONAL and present", structType.Field(i).Name, structType))
				}
			}
			// for open type reference
			if structParams[i].openType {
				fieldName := structParams[i].referenceFieldName
				var index int
				for index = 0; index < i; index++ {
					if structType.Field(index).Name == fieldName {
						break
					}
				}
				if index == i {
					return fmt.Errorf("Open type is not reference to the other field in the struct")
				}
				structParams[i].referenceFieldValue = new(int64)
				if referenceFieldValue, err := getReferenceFieldValue(val.Field(index)); err != nil {
					return err
				} else {
					*structParams[i].referenceFieldValue = referenceFieldValue
				}
			}
            
            if _, ok := val.Field(i).Interface().(AperDecoder); ok {
                
                decoderType := reflect.New(val.Field(i).Type().Elem())
                val.Field(i).Set(decoderType)
                adInterface := val.Field(i).Interface().(AperDecoder)
                err := adInterface.AperDecode(pd, structParams[i] )
                if err!=nil{
                    return err
                }
                continue
            } 


            fieldName := val.Type().Field(i).Name
            _, isCustom := CustomFieldValues[fieldName]
            if isCustom {
                fieldVal, err := CustomFieldValues[fieldName](val, i, fieldName)
                if err != nil {
                    return err
                }
                field := val.Field(i)
                if field.Kind() == reflect.String {
                    field.SetString(fieldVal.String())
                } else if field.Kind() == reflect.Int64{
                    field.SetInt(fieldVal.Int())
                } else if field.Kind() == reflect.Bool {
                    field.SetBool(fieldVal.Bool())
                } else {
				    err = fmt.Errorf("Custom type returned is neither string nor int64")
                }

            } else {
                /*
                fmt.Println("\n\t-----struct-----\n")
                spew.Dump(val.Interface())
                fmt.Println("\t--field--")
                spew.Dump(val.Field(i).Interface())
                fmt.Println("\t--sp--")
                spew.Dump(structParams[i])
                fmt.Println("\n\t----pd---\n")
                spew.Dump(pd)
                fmt.Println("\n\t----------\n")
                */

                if err := ParseField(val.Field(i), pd, structParams[i]); err != nil {
                    return err
                }
                
            }
		}
		
        return nil

	case reflect.Slice:
		sliceType := fieldType
		if newSlice, err := pd.parseSequenceOf(sizeExtensible, params, sliceType); err != nil {
			return err
		} else {
			val.Set(newSlice)
			return nil
		}
	case reflect.String:
		perTrace(2, "Decoding PrintableString using Octet String decoding method")

		if octetString, err := pd.ParseOctetString(sizeExtensible, params.sizeLowerBound, params.sizeUpperBound); err != nil {
			return err
		} else {
			printableString := string(octetString.Bytes)
			val.SetString(printableString)
			perTrace(2, fmt.Sprintf("Decoded PrintableString : \"%s\"", printableString))
			return nil
		}
	}
	return fmt.Errorf("unsupported: " + v.Type().String())
}

// Unmarshal parses the APER-encoded ASN.1 data structure b
// and uses the reflect package to fill in an arbitrary value pointed at by value.
// Because Unmarshal uses the reflect package, the structs
// being written to must use upper case field names.
//
// An ASN.1 INTEGER can be written to an int, int32, int64,
// If the encoded value does not fit in the Go type,
// Unmarshal returns a parse error.
//
// An ASN.1 BIT STRING can be written to a BitString.
//
// An ASN.1 OCTET STRING can be written to a []byte.
//
// An ASN.1 OBJECT IDENTIFIER can be written to an
// ObjectIdentifier.
//
// An ASN.1 ENUMERATED can be written to an Enumerated.
//
// Any of the above ASN.1 values can be written to an interface{}.
// The value stored in the interface has the corresponding Go type.
// For integers, that type is int64.
//
// An ASN.1 SEQUENCE OF x can be written
// to a slice if an x can be written to the slice's element type.
//
// An ASN.1 SEQUENCE can be written to a struct
// if each of the elements in the sequence can be
// written to the corresponding element in the struct.
//
// The following tags on struct fields have special meaning to Unmarshal:
//
//	optional        	OPTIONAL tag in SEQUENCE
//	sizeExt             specifies that size  is extensible
//	valueExt            specifies that value is extensible
//	sizeLB		        set the minimum value of size constraint
//	sizeUB              set the maximum value of value constraint
//	valueLB		        set the minimum value of size constraint
//	valueUB             set the maximum value of value constraint
//	default             sets the default value
//	openType            specifies the open Type
//  referenceFieldName	the string of the reference field for this type (only if openType used)
//  referenceFieldValue	the corresponding value of the reference field for this type (only if openType used)
//
// Other ASN.1 types are not supported; if it encounters them,
// Unmarshal returns a parse error.
func Unmarshal(b []byte, value interface{}) error {
	return UnmarshalWithParams(b, value, "")
}

func make_human_readable(value interface{}) error{
     
    imJSON, err := json.MarshalIndent(value,"","    ")
    if err!=nil {
        return fmt.Errorf("Failed to Marshal JSON")
    }
    fmt.Println(string(len(imJSON)))
    //spew.Dump(imJSON)
    return nil

}

// UnmarshalWithParams allows field parameters to be specified for the
// top-level element. The form of the params is the same as the field tags.
func UnmarshalWithParams(b []byte, value interface{}, params string) error {
	v := reflect.ValueOf(value).Elem()
	pd := &PerBitData{b, 0, 0}
    ret := ParseField(v, pd, parseFieldParameters(params))
   
    return(ret)
}
