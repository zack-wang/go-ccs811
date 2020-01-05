//--------------------------------------------------------------------------------------------------
//
// Copyright (c) 2020 Zack Wang
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
  "math"
	//"bytes"
	//"encoding/binary"
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
  InterruptMode byte=0
  InterruptThreshold byte=0
  SamplingRate byte=  CCS811_DRIVE_MODE_10SEC
  Dev *i2c.Device

  //// Private Variables
	b1=make([]byte,1)
	b2=make([]byte,2)
	b3=make([]byte,3)
	b4=make([]byte,4)
  b8=make([]byte,8)
)

func getStatus(d *i2c.Device) (byte,error) {
    err:=d.ReadReg(CCS811_STATUS,b1)
	  if err != nil {
		    return 0,err
	  }else{
        return b1[0],nil
    }
}

func setReset(d *i2c.Device) {
  	resetcode:=[]byte {0x11, 0xE5, 0x72, 0x8A}
  	d.WriteReg(CCS811_SW_RESET,resetcode)
}

func verifyId(d *i2c.Device) bool{
  	err:=d.ReadReg(CCS811_HW_ID,b1)
    if err==nil && b1[0]==CCS811_HW_ID_CODE{
      return true
    }else{
      return false
    }
}

func setConfig(d *i2c.Device){
  bin1:=0x01 & InterruptThreshold
  bin2:=0x01 & InterruptMode
  bin3:=0x03 & SamplingRate
  b1[0]= bin1 <<2 | bin2 <<3 | bin3<< 4
  d.WriteReg(CCS811_MEAS_MODE,b1)
}

func isDataReady(d *i2c.Device) bool{
  sts,err:=getStatus(d)
  if err==nil{
    return (sts & DATA_READY_BIT) != 0
  }else{
    log.Println("isCCS811 DateReady Error=",err)
    return false
  }
}

func setEnv(d *i2c.Device, temperature float64, humidity int){
  /* Humidity is stored as an unsigned 16 bits in 1/512%RH. The
	default value is 50% = 0x64, 0x00. As an example 48.5%
	humidity would be 0x61, 0x00.*/

	/* Temperature is stored as an unsigned 16 bits integer in 1/512
	degrees; there is an offset: 0 maps to -25°C. The default value is
	25°C = 0x64, 0x00. As an example 23.5% temperature would be
	0x61, 0x00.
	The internal algorithm uses these values (or default values if
	not set by the application) to compensate for changes in
	relative humidity and ambient temperature.*/

  h:=humidity<<1 // H=H*2

  // -25 = 0 ,+25=100,+75=200,+125=300
  t:=math.Round(temperature*2.0)/2.0
  ut:= uint32( math.Round( (t +25.0) / 25.0 * 100.0 ) )
	b:=[]byte { byte(h), 0x00, byte( ut & 0xFF), byte( (ut>>8) & 0xFF )}

	d.WriteReg(CCS811_ENV_DATA, b)

}
func Begin(d string,a int) bool{
  Dev,err:=i2c.Open(&i2c.Devfs{Dev:d},a)
  defer Dev.Close()

  if err!=nil{
    log.Println("I2C BUS Error")
    return false
  }else{

    setReset(Dev)
    time.Sleep(100 * time.Millisecond)
    tf:=verifyId(Dev)
    if tf {
      log.Println("is CCS811")
    }else{
      return false
    }
    //// Start APP
    Dev.Write([]byte{byte(CCS811_BOOTLOADER_APP_START),})
    time.Sleep(100 * time.Millisecond)
    sts,err:=getStatus(Dev)
    if err==nil{
      if sts & ERROR_BIT !=0{
        log.Println("CCS811 device has error")
        return false
      }
  	  if sts & FW_MODE_BIT ==0{
        log.Println("In FW mode")
        return false
      }
    }
    setConfig(Dev)
    log.Println("OK")
    return true
  }
}

func ReadData(d string, a int) (uint16,uint16,bool){
  sensor,err:=i2c.Open(&i2c.Devfs{Dev:d},a)
  var eCO2, TVOC uint16
  eCO2=0
  TVOC=0
  isValid:=false
  defer sensor.Close()

  if err!=nil{
    log.Println("I2C Error-",err)
    return 0,0,isValid
  }else{
    if isDataReady(sensor){
      //log.Println("Data IS ready")
      isValid=true
    }else{
      //log.Println("Data is NOT ready")
      isValid=false
    }
    //// Data is Ready
    err=sensor.ReadReg(CCS811_ALG_RESULT_DATA, b8)
    if err==nil{
  	    eCO2 = ( uint16(b8[0]) << 8) | uint16(b8[1])
  	    TVOC = ( uint16(b8[2]) << 8) | uint16(b8[3])
    }else{
        log.Println("I2C Error-",err)
        isValid=false
    }

  }
  return eCO2,TVOC,isValid
}
