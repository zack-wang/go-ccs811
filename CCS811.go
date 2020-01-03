//--------------------------------------------------------------------------------------------------
//
// Copyright (c) 2020 zack Wang
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
// associated documentation files (the "Software"), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial
// portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
//--------------------------------------------------------------------------------------------------

package ccs811
import (
  "log"
	"bytes"
	"encoding/binary"
	"time"
	"golang.org/x/exp/io/i2c"
)
const(
  //// Registers
  CCS811_STATUS = 0x00
  CCS811_MEAS_MODE = 0x01
  CCS811_ALG_RESULT_DATA = 0x02
  CCS811_RAW_DATA = 0x03
  CCS811_ENV_DATA = 0x05
  CCS811_NTC = 0x06
  CCS811_THRESHOLDS = 0x10
  CCS811_BASELINE = 0x11
  CCS811_HW_ID = 0x20
  CCS811_HW_VERSION = 0x21
  CCS811_FW_BOOT_VERSION = 0x23
  CCS811_FW_APP_VERSION = 0x24
  CCS811_ERROR_ID = 0xE0
  CCS811_SW_RESET = 0xFF
  //// Bootloader Registers
  CCS811_BOOTLOADER_APP_ERASE = 0xF1
  CCS811_BOOTLOADER_APP_DATA = 0xF2
  CCS811_BOOTLOADER_APP_VERIFY = 0xF3
  CCS811_BOOTLOADER_APP_START = 0xF4
  //// Drive mode
  CCS811_DRIVE_MODE_IDLE = 0x00
  CCS811_DRIVE_MODE_1SEC = 0x01
  CCS811_DRIVE_MODE_10SEC = 0x02
  CCS811_DRIVE_MODE_60SEC = 0x03
  CCS811_DRIVE_MODE_250MS = 0x04
  //// CONSTANTs
  CCS811_HW_ID_CODE	=	0x81
  CCS811_REF_RESISTOR	=	100000
  //// STATUS - Bitwise
  ERROR_BIT = 0x01
  DATA_READY_BIT = 0x08
  APP_VALID_BIT = 0x10
  FW_MODE_BIT = 0x80
  //// ERROR - Bitwise
  WRITE_REG_INVALID = 0x01
  READ_REG_INVALID = 0x02
  MEASMODE_INVALID = 0x04
  MAX_RESISTANCE = 0x08
  HEATER_FAULT = 0x10
  HEATER_SUPPLY = 0x20
)

var (
  //// Private Variables
	b1=make([]byte,1)
	b2=make([]byte,2)
	b3=make([]byte,3)
	b4=make([]byte,4)
  b8=make([]byte,8)
)

func getStatus(d string, a int) (byte,error) {
  i2c0,err:=i2c.Open(&i2c.Devfs{Dev:d},a)
  defer i2c0.Close()

  if err!=nil{
    log.Println("I2C Error")
    return 0,err
  }else{
    err = i2c0.ReadReg(CCS811_STATUS,b1)
	  if err != nil {
		    return 0,err
	  }else{
        return b1[0],nil
    }
  }
}

func setReset(d string,a int) {
  i2c0,err:=i2c.Open(&i2c.Devfs{Dev:d},a)
  defer i2c0.Close()

  if err!=nil{
    log.Println("I2C Error")
    return
  }else{
  	resetcode:=[]byte {0x11, 0xE5, 0x72, 0x8A}
  	i2c0.WriteReg(CCS811_SW_RESET,resetcode)
  }
}

func verifyId(d string,a int) bool{
  i2c0,err:=i2c.Open(&i2c.Devfs{Dev:d},a)
  defer i2c0.Close()

  if err!=nil{
    log.Println("I2C Error")
    return false
  }else{
  	err=i2c0.ReadReg(CCS811_HW_ID,b1)
    if err==nil && b1[0]==CCS811_HW_ID_CODE{
      return true
    }else{
      return false
    }
  }
}

func Begin(d string,a int) bool{
  setReset(d,a)
  time.Sleep(100 * time.Millisecond)
  tf:=verifyId(d,a)
  if tf {
    log.Println("is CCS811")
    return true
  }else{
    return false
  }
}
