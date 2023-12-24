/*
 * Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this software
 * and associated documentation files (the “Software”), to deal in the Software without restriction,
 * including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial
 * portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
 * LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
 * IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
 * WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
 * SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package webhooks

import (
	"fmt"
	"os/exec"
	"strings"
)

const listIPAddressesCmd = `whois -h whois.radb.net — '-i origin AS32934' | grep ^route | awk '{print $2}' | sort`

// ListIPAddresses returns a list of IP addresses that you can use to allow-list our webhook
// servers in your firewall or network configuration.
//
// You can get the IP addresses of our webhook servers by running the following command in
// your terminal:
//
//	whois -h whois.radb.net — '-i origin AS32934' | grep ^route | awk '{print $2}' | sort
//
// We periodically change these IP addresses so if you are allow-listing our servers you may
// want to occasionally regenerate this list and update your allow-list accordingly.
func ListIPAddresses() ([]string, error) {
	cmd := exec.Command("bash", "-c", listIPAddressesCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list IP addresses: %w", err)
	}

	routes := strings.Split(strings.TrimSpace(string(output)), "\n")

	return routes, nil
}
