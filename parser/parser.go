package parser

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type StopComplex struct {
	ID       string
	Name     string
	CityID   string
	CityName string
}

type Parser struct {
	OnStopComplexesParsed func([]StopComplex) error

	r *bufio.Reader
}

const fileTerminator = "####"

func NewParser(r io.Reader) *Parser {
	return &Parser{
		r: bufio.NewReader(r),
	}
}

func (p *Parser) Parse() error {
	return p.ScanSections()
}

func (p *Parser) readLine() (string, error) {
	bytes, err := p.r.ReadBytes('\n')
	if err != nil {
		return "", err
	}

	str := strings.TrimRight(string(bytes), "\r\n")

	return str, nil
}

func (p *Parser) scanSectionHeader() (string, int, error) {
	line, err := p.readLine()
	if err != nil {
		return "", 0, err
	}

	if line == fileTerminator {
		return "", 0, io.EOF
	}

	r := regexp.MustCompile(`^\*([A-Z]{2})\s*(\d+)$`)
	matches := r.FindStringSubmatch(line)
	if len(matches) != 3 {
		return "", 0, errors.New("wrong section header: " + line)
	}

	length, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", 0, errors.Wrap(err, "wrong section length: "+line)
	}

	return matches[1], length, nil
}

var missingSectionTerminatorErr = errors.New("missing section terminator")

func (p *Parser) scanSectionTerminator(sectionHeader string) error {
	last, err := p.readLine()
	if err != nil {
		return err
	}
	if last != "#"+sectionHeader {
		return missingSectionTerminatorErr
	}

	return nil
}

const (
	runningDayTypesHeader            = "TY"
	runningDayTypesCalendarHeader    = "KA"
	runningDayTypeLineCalendarHeader = "KD"
	stopComplexesHeader              = "ZA"
	stopsHeader                      = "ZP"
	citiesHeader                     = "SM"
	routesSchedulesHeader            = "LL"
)

func (p *Parser) ScanSections() error {
	for {
		section, length, err := p.scanSectionHeader()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch section {
		case runningDayTypesHeader:
			if err := p.ScanRunningDayTypes(length); err != nil {
				return err
			}
		case runningDayTypesCalendarHeader:
			if err := p.ScanRunningDayTypesCalendar(length); err != nil {
				return err
			}
		case runningDayTypeLineCalendarHeader:
			if err := p.ScanRunningDayTypeLineCalendar(length); err != nil {
				return err
			}
		case stopComplexesHeader:
			if err := p.ScanStopComplexes(length); err != nil {
				return err
			}
		case stopsHeader:
			return nil
		case citiesHeader:
		case routesSchedulesHeader:
		default:
			return errors.New("unknown section: " + section)
		}
	}
}

func (p *Parser) ScanRunningDayTypes(length int) error {
	for i := 0; i < length; i++ {
		p.readLine() // todo
	}

	return p.scanSectionTerminator(runningDayTypesHeader)
}

func (p *Parser) ScanRunningDayTypesCalendar(length int) error {
	for i := 0; i < length; i++ {
		p.readLine() // todo
	}

	return p.scanSectionTerminator(runningDayTypesCalendarHeader)
}

func (p *Parser) ScanRunningDayTypeLineCalendar(length int) error {
	for {
		err := p.scanSectionTerminator(runningDayTypeLineCalendarHeader)
		if err == nil {
			break
		}
		if err != missingSectionTerminatorErr {
			return err
		}
	}

	return nil
}

func (p *Parser) ScanStopComplexes(length int) error {
	r := regexp.MustCompile(`(\d{4})\s+(.{30})\s+(.{2})\s+(.+)$`)

	stopComplexes := make([]StopComplex, length)
	for i := 0; i < length; i++ {
		line, err := p.readLine()
		if err != nil {
			return err
		}

		matches := r.FindStringSubmatch(line)
		stopComplexes[i] = StopComplex{
			ID:     matches[1],
			Name:   strings.TrimRight(matches[2], " ,"),
			CityID: matches[3],
			//lint:ignore SA1019 this will work for us
			CityName: strings.Title(strings.ToLower(matches[4])),
		}
	}

	if p.OnStopComplexesParsed != nil {
		err := p.OnStopComplexesParsed(stopComplexes)
		if err != nil {
			return err
		}
	}

	return p.scanSectionTerminator(stopComplexesHeader)
}
