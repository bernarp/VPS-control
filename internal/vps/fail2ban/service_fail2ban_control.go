package fail2ban

import (
	"DiscordBotControl/internal/apierror"
	"DiscordBotControl/internal/vps"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type ControlService struct {
	base   *vps.BaseVpsService
	logger *zap.Logger
}

func NewControlService(
	base *vps.BaseVpsService,
	logger *zap.Logger,
) *ControlService {
	return &ControlService{
		base:   base,
		logger: logger.Named("fail2ban_service"),
	}
}

func (s *ControlService) GetGlobalStatus() (*Fail2BanStatusDTO, error) {
	out, err := exec.Command(CmdSudo, CmdFail2Ban, ArgStatus).CombinedOutput()
	if err != nil {
		return nil, apierror.Errors.FAIL2BAN_EXECUTION_ERROR.Wrap(fmt.Errorf("%w: %s", err, string(out)))
	}

	strOut := string(out)
	res := &Fail2BanStatusDTO{JailList: []string{}}

	jailListRegex := regexp.MustCompile(ReJailList)
	matches := jailListRegex.FindStringSubmatch(strOut)
	if len(matches) > 1 {
		jails := strings.Split(matches[1], ",")
		for _, j := range jails {
			name := strings.TrimSpace(j)
			if name != "" {
				res.JailList = append(res.JailList, name)
			}
		}
	}
	res.JailCount = len(res.JailList)

	return res, nil
}

func (s *ControlService) GetJailDetails(jailName string) (*JailDetailsDTO, error) {
	out, err := exec.Command(CmdSudo, CmdFail2Ban, ArgStatus, jailName).CombinedOutput()
	if err != nil {
		output := string(out)
		if strings.Contains(output, ErrOutputDoesNotExist) || strings.Contains(output, ErrOutputNotFound) {
			return nil, apierror.Errors.FAIL2BAN_JAIL_NOT_FOUND
		}
		return nil, apierror.Errors.FAIL2BAN_EXECUTION_ERROR.Wrap(fmt.Errorf("%w: %s", err, output))
	}

	strOut := string(out)
	res := &JailDetailsDTO{
		JailName:     jailName,
		BannedIPList: []string{},
	}

	res.CurrentlyFailed = s.parseIntField(strOut, ReCurrentlyFailed)
	res.TotalFailed = s.parseIntField(strOut, ReTotalFailed)
	res.CurrentlyBanned = s.parseIntField(strOut, ReCurrentlyBanned)
	res.TotalBanned = s.parseIntField(strOut, ReTotalBanned)

	// Парсим список IP (может занимать несколько строк)
	ipListRegex := regexp.MustCompile(ReBannedIPList)
	ipMatches := ipListRegex.FindStringSubmatch(strOut)
	if len(ipMatches) > 1 {
		rawIps := strings.TrimSpace(ipMatches[1])
		if rawIps != "" {
			res.BannedIPList = strings.Fields(rawIps)
		}
	}

	return res, nil
}

func (s *ControlService) UnbanIP(jail, ip string) error {
	out, err := exec.Command(CmdSudo, CmdFail2Ban, ArgSet, jail, ArgUnbanIP, ip).CombinedOutput()
	if err != nil {
		output := string(out)
		if strings.Contains(output, ErrOutputIsNotBanned) {
			return apierror.Errors.FAIL2BAN_IP_NOT_BANNED
		}
		if strings.Contains(output, ErrOutputJailNotFound) || strings.Contains(output, ErrOutputDoesNotExist) {
			return apierror.Errors.FAIL2BAN_JAIL_NOT_FOUND
		}
		return apierror.Errors.FAIL2BAN_EXECUTION_ERROR.Wrap(fmt.Errorf("%w: %s", err, output))
	}

	s.logger.Info("IP unbanned successfully", zap.String("jail", jail), zap.String("ip", ip))
	return nil
}

func (s *ControlService) parseIntField(input, pattern string) int {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		val, _ := strconv.Atoi(matches[1])
		return val
	}
	return 0
}
