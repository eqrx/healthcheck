// Copyright (C) 2021 Alexander Sowitzki
//
// This program is free software: you can redistribute it and/or modify it under the terms of the
// GNU Affero General Public License as published by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
// warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for more
// details.
//
// You should have received a copy of the GNU Affero General Public License along with this program.
// If not, see <https://www.gnu.org/licenses/>.

package internal

import (
	"errors"
	"fmt"
	"os"
)

const (
	// EnvNameService defines the name of the environment variable that contains the type of the service this run checks.
	// It must be one of the defined ServiceType values.
	EnvNameService = "HEALTHCHECK_SERVICE"
	// EnvNameTarget defines the name of the environment variable that contains the target of this check. What that means
	// depends on the given ServiceType.
	EnvNameTarget = "HEALTHCHECK_TARGET"
	// EnvNamePing defines the name of the environment variable that contains the healthchecks.io check UUID that shall be
	// triggered if the check succeeds.
	EnvNamePing = "HEALTHCHECK_PING"
)

// ServiceType defines what kind of service is tested by this run. See constants in this package for more explanation.
type ServiceType string

const (
	// ServiceSMTPv6 defines that a SMTP service listening on a IPv6 address shall be checked. It does this by
	// resolving the MX record for the domain specified the EnvNameTarget env, connecting to the first IPv6 result
	// and asking the server to switch over to TLS via STARTTLS. If this succeeds and the provided TLS certificate is
	// valid and trusted by the system healthcheck runs on, the test is considered successful.
	ServiceSMTPv6 ServiceType = "smtpv6"
	// ServiceSMTPv4 is the same as ServiceSMTPv6 but only accepts IPv4 addresses in the MX record.
	ServiceSMTPv4 ServiceType = "smtpv4"
	// ServiceMatrixv6 defines that a matrix service listening on a IPv6 address shall be checked. It does this by
	// resolving the SRV record for the domain specified the EnvNameTarget env, connecting to the first IPv6 result
	// via HTTPS and querying the synapse health endpoint. If the endpoint returns with HTTP code 200 the test is
	// considered successful.
	ServiceMatrixv6 ServiceType = "matrixv6"
	// ServiceMatrixv4 is the same as ServiceMatrixv6 but only accepts IPv6 addresses in the SRV record.
	ServiceMatrixv4 ServiceType = "matrixv4"
	// ServiceCeph defines that a ceph cluster shall be checked. It does this by spawning the command
	// `ceph status -f json` via the exec module and captures the output. If ceph reports "HEALTH_OK" as health status
	// the test is considered successful. This requires a working ceph setup. A broken setup is considered a check failure.
	ServiceCeph ServiceType = "ceph"
)

// env wraps all required configuration items together for easier handling.
type env struct {
	key     string
	service ServiceType
	target  string
}

var (
	// errEnvMissing is raised when an environment variable is missing.
	errEnvMissing = errors.New("environment variable is not set")
	// errUnknownService is raised when an unknown service was specified via the EnvNameService variable.
	errUnknownService = errors.New("service not known")
)

// readEnv looks up all configuration items and returns them as env instance. If an item is missing it returns and
// error and no env. Values are not checked for validity here.
func readEnv() (*env, error) {
	service, serviceOK := os.LookupEnv(EnvNameService)
	if !serviceOK {
		return nil, fmt.Errorf("%w: %s", errEnvMissing, EnvNameService)
	}

	key, keyOK := os.LookupEnv(EnvNamePing)
	if !keyOK {
		return nil, fmt.Errorf("%w: %s", errEnvMissing, EnvNamePing)
	}

	target, targetOK := os.LookupEnv(EnvNameTarget)
	if !targetOK {
		return nil, fmt.Errorf("%w: %s", errEnvMissing, EnvNameTarget)
	}

	return &env{key, ServiceType(service), target}, nil
}
