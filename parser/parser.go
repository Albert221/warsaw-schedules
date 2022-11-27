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

type Stop struct {
	StopComplexID string
	StopID        string
	Street        string
	Direction     string
	Latitude      *float64
	Longitude     *float64
	Platform      *int
}

type Parser struct {
	OnStopComplexesParsed func([]StopComplex) error
	OnStopsParsed         func([]Stop) error

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

	r := regexp.MustCompile(`\*([A-Z]{2})\s*(\d+)$`)
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
	if !strings.Contains(last, "#"+sectionHeader) {
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
			if err := p.ScanStops(length); err != nil {
				return err
			}
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

var stopComplexRowRegexp = regexp.MustCompile(`(\d{4})\s+(.{30})\s+(.{2})\s+(.+)$`)

func (p *Parser) ScanStopComplexes(length int) error {
	stopComplexes := make([]StopComplex, length)
	for i := 0; i < length; i++ {
		line, err := p.readLine()
		if err != nil {
			return err
		}

		matches := stopComplexRowRegexp.FindStringSubmatch(line)
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

func (p *Parser) ScanStops(length int) error {
	const stopsRowsHeader = "PR"
	stopRegexp := regexp.MustCompile(`(\d{4})(\d{2})\s+(\d+)\s+Ul\.\/Pl\.: (.+?)Kier\.: (.+)Y=(.+)\s+X=(.+) Pu=(\d+|\?)$`)

	var stops []Stop
	for i := 0; i < length; i++ {
		if _, err := p.readLine(); err != nil {
			return err
		}

		section, length, err := p.scanSectionHeader()
		if err != nil {
			return err
		}
		if section != stopsRowsHeader {
			return errors.New("wrong section: " + section)
		}

		for j := 0; j < length; j++ {
			line, err := p.readLine()
			if err != nil {
				return err
			}

			matches := stopRegexp.FindStringSubmatch(line)
			if len(matches) != 9 {
				return errors.New("wrong stop: " + line)
			}

			length, err := strconv.Atoi(matches[3])
			if err != nil {
				return errors.Wrap(err, "wrong stop length: "+line)
			}
			for k := 0; k < length; k++ {
				if _, err := p.readLine(); err != nil {
					return err
				}
			}

			lat := parseNullableFloat(strings.TrimSpace(matches[6]))
			lng := parseNullableFloat(strings.TrimSpace(matches[7]))
			platform := parseNullableInt(matches[8])

			stop := Stop{
				StopComplexID: matches[1],
				StopID:        matches[2],
				Street:        strings.TrimRight(matches[4], ", "),
				Direction:     strings.TrimRight(matches[5], ", "),
				Latitude:      lat,
				Longitude:     lng,
				Platform:      platform,
			}
			stops = append(stops, stop)
		}

		if err := p.scanSectionTerminator(stopsRowsHeader); err != nil {
			return err
		}
	}

	if p.OnStopsParsed != nil {
		err := p.OnStopsParsed(stops)
		if err != nil {
			return err
		}
	}

	return p.scanSectionTerminator(stopsHeader)
}

func parseNullableFloat(s string) *float64 {
	parsed, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}

	return &parsed
}

func parseNullableInt(s string) *int {
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}

	return &parsed
}
