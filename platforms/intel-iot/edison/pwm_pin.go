package edison

import (
	"fmt"
	"strconv"
)

// pwmPath returns pwm base path
func pwmPath() string {
	return "/sys/class/pwm/pwmchip0"
}

// pwmExportPath returns export path
func pwmExportPath() string {
	return pwmPath() + "/export"
}

// pwmUnExportPath returns unexport path
func pwmUnExportPath() string {
	return pwmPath() + "/unexport"
}

// pwmDutyCyclePath returns duty_cycle path for specified pin
func pwmDutyCyclePath(pin int) string {
	return fmt.Sprintf("%s/pwm%d/duty_cycle", pwmPath(), pin)
}

// pwmPeriodPath returns period path for specified pin
func pwmPeriodPath(pin int) string {
	return fmt.Sprintf("%s/pwm%d/period", pwmPath(), pin)
}

// pwmEnablePath returns enable path for specified pin
func pwmEnablePath(pin int) string {
	return fmt.Sprintf("%s/pwm%d/enable", pwmPath(), pin)
}

type pwmPin struct {
	pin int
}

// newPwmPin returns an exported and enabled pwmPin
func newPwmPin(pin int) *pwmPin {
	return &pwmPin{pin: pin}
}

// enable writes value to pwm enable path
func (p *pwmPin) enable(v bool) (err error) {
	val := "0"
	if v {
		val = "1"
	}
	_, err = writeFile(pwmEnablePath(p.pin), []byte(val))
	return
}

// period reads from pwm period path and returns value
func (p *pwmPin) period() (period int, err error) {
	buf, err := readFile(pwmPeriodPath(p.pin))
	if err != nil {
		return
	}
	period, err = strconv.Atoi(string(buf[0 : len(buf)-1]))
	if err != nil {
		return 0, err
	}
	return period, nil
}

// duty reads from pwm duty path and returns value
func (p *pwmPin) duty() (duty int, err error) {
	buf, err := readFile(pwmDutyCyclePath(p.pin))
	if err != nil {
		return
	}
	duty, err = strconv.Atoi(string(buf[0 : len(buf)-1]))
	if err != nil {
		return 0, err
	}
	return duty, nil
}

// pwmWrite writes the period and duty to the appropriate paths;
// it will handle writing them in the correct order
func (p *pwmPin) pwmWrite(period, duty int) (err error) {
	if duty > period {
		return fmt.Errorf("duty cannot be greater than period: %d > %d", duty, period)
	}

	// Get current duty
	curDuty, err := p.duty()
	if err != nil {
		return err
	}

	// Need to write in certain order:
	// | Pre-condition                        | Action                                                                        |
	// | period >= curDuty, duty <= curPeriod | don't care                                                                    |
	// | period >= curDuty, duty > curPeriod  | write period, then duty                                                       |
	// | period < curDuty, duty <= curPeriod  | write duty, then period                                                       |
	// | period < curDuty, duty > curPeriod   | not possible, CP < D <= P < CD, would not allow current period < current duty |
	if period >= curDuty {
		_, err = writeFile(pwmPeriodPath(p.pin), []byte(strconv.Itoa(period)))
		if err == nil {
			_, err = writeFile(pwmDutyCyclePath(p.pin), []byte(strconv.Itoa(duty)))
		}
	} else {
		_, err = writeFile(pwmDutyCyclePath(p.pin), []byte(strconv.Itoa(duty)))
		if err == nil {
			_, err = writeFile(pwmPeriodPath(p.pin), []byte(strconv.Itoa(period)))
		}
	}
	return
}

// export writes pin to pwm export path
func (p *pwmPin) export() (err error) {
	_, err = writeFile(pwmExportPath(), []byte(strconv.Itoa(p.pin)))
	return
}

// export writes pin to pwm unexport path
func (p *pwmPin) unexport() (err error) {
	_, err = writeFile(pwmUnExportPath(), []byte(strconv.Itoa(p.pin)))
	return
}
