// MIT License
//
// (C) Copyright [2018-2021] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	base "github.com/Cray-HPE/hms-base"
	"github.com/Cray-HPE/hms-smd/pkg/sm"
)

type Response struct {
	C int    `json:"code"`
	M string `json:"message"`
}

// Send a JSON error or message.  If the ecode indicates an error, send
// a properly formatted RFC7807 problem.
// If it does not, fall back to the original CAPMC format, which will
// now just be used for success cases or odd HTTP status codes that
// don't suggest a RFC7807 problem response.
// We use the 7807 problem format for 4xx and 5xx errors, though
// in practice the latter (server errors) will probably not be used here
// as they do not describe invalid requests but server-specific issues.
func sendJsonError(w http.ResponseWriter, ecode int, message string) {
	if ecode < 400 {
		sendJsonResponse(w, ecode, message)
	} else {
		// Use library function in HMS base.  Problem will be
		// a generic one with title matching the HTTP Status code text
		// with message as the details field.  For this type of problem
		// title can just be set to "about:blank" so we need no
		// custom URL.  The optional instance field is omitted as well
		// so no URL/URI is needed there either.
		base.SendProblemDetailsGeneric(w, ecode, message)
	}
}

// Send a simple message for cases where need a non-error response.  If
// a more feature filled message needs to be returned then do it with a
// different function.  Code is the http status response, converted to
// zero for success-related responses.
func sendJsonResponse(w http.ResponseWriter, ecode int, message string) {
	// if the HTTP call was a success then put zero in the returned json
	// error field. This is what capmc does.
	http_code := ecode
	if ecode >= 200 && ecode <= 299 {
		ecode = 0
	}
	data := Response{ecode, message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if http_code != http.StatusNoContent {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			fmt.Printf("Yikes, I couldn't encode a JSON status response: %s\n", err)
		}
	}
}

// If an HMS-created error, return it as a bad status request.  If another
// type, just print a generic failed to query DB error message.
func sendJsonDBError(w http.ResponseWriter, prefix, internalErr string, err error) {
	if base.IsHMSError(err) {
		sendJsonError(w, http.StatusBadRequest, prefix+err.Error())
	} else {
		if internalErr != "" {
			sendJsonError(w, http.StatusInternalServerError, internalErr)
		} else {
			sendJsonError(w, http.StatusInternalServerError,
				"failed to query DB.")
		}
	}
}

// When creating one or more resources, we want return 201 Created along with
// an array of links to the newly-created resources.  The location header
// should also be set to the new resource (if just one) or to the affected
// collection.
func sendJsonNewResourceIDArray(w http.ResponseWriter, collectionURI string,
	uris []*sm.ResourceURI) {

	http_code := 201
	w.Header().Set("Content-Type", "application/json")

	numUris := len(uris)
	if numUris == 0 {
		http_code = 204
	} else if numUris == 1 {
		// Set Location to most specific component affected.  If only
		// a single new resource is created, use its URI
		w.Header().Set("Location", uris[0].URI)
	} else {
		// Otherwise, set it to the collection.
		w.Header().Set("Location", collectionURI)
	}
	w.WriteHeader(http_code)
	if numUris != 0 {
		err := json.NewEncoder(w).Encode(uris)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// When creating a single resource, we want return 201 Created along with
// a link to the newly-created resource.  The location header
// should also be set to the new resource.
func sendJsonNewResourceID(w http.ResponseWriter, uri *sm.ResourceURI) {
	http_code := 201
	w.Header().Set("Content-Type", "application/json")

	if uri == nil {
		http_code = 204
	} else {
		// Set Location to the newly-created object's URI
		w.Header().Set("Location", uri.URI)
	}
	w.WriteHeader(http_code)
	if uri != nil {
		err := json.NewEncoder(w).Encode(uri)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Return a link to a single resource that is not new.
func sendJsonResourceID(w http.ResponseWriter, uri *sm.ResourceURI) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")

	if uri == nil {
		http_code = 204
	}
	w.WriteHeader(http_code)
	if uri != nil {
		err := json.NewEncoder(w).Encode(uri)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Return a link to an array of resources that are not new.
func sendJsonResourceIDArray(w http.ResponseWriter, uris []*sm.ResourceURI) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")

	if len(uris) == 0 {
		http_code = 204
	}
	w.WriteHeader(http_code)
	if len(uris) != 0 {
		err := json.NewEncoder(w).Encode(uris)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// HMS Component response
func sendJsonCompRsp(w http.ResponseWriter, comp *base.Component) {
	http_code := 200
	if comp == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if comp != nil {
		err := json.NewEncoder(w).Encode(comp)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Named HMS Component array response (i.e. GET on HMS Components collection)
func sendJsonCompArrayRsp(w http.ResponseWriter, comps *base.ComponentArray) {
	http_code := 200
	if comps == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if comps != nil {
		err := json.NewEncoder(w).Encode(comps)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Node NID Mapping response, single entry
func sendJsonNodeMapRsp(w http.ResponseWriter, m *sm.NodeMap) {
	http_code := 200
	if m == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if m != nil {
		err := json.NewEncoder(w).Encode(m)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Named NodeMaps array response (i.e. GET on NodeMaps collection)
func sendJsonNodeMapArrayRsp(
	w http.ResponseWriter,
	nnms *sm.NodeMapArray) {

	http_code := 200
	if nnms == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if nnms != nil {
		err := json.NewEncoder(w).Encode(nnms)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Hardware inventory by Location, single entry, with FRU data if populated
func sendJsonHWInvByLocRsp(w http.ResponseWriter, hl *sm.HWInvByLoc) {
	http_code := 200
	if hl == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if hl != nil {
		err := json.NewEncoder(w).Encode(hl)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Hardware inventory by Location, multiple entry, with FRU data if populated
func sendJsonHWInvByLocsRsp(w http.ResponseWriter, hl []*sm.HWInvByLoc) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if hl != nil {
		err := json.NewEncoder(w).Encode(hl)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Hardware inventory by FRU, single entry, by FRU ID
func sendJsonHWInvByFRURsp(w http.ResponseWriter, hf *sm.HWInvByFRU) {
	http_code := 200
	if hf == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if hf != nil {
		err := json.NewEncoder(w).Encode(hf)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Hardware inventory by FRU ID, multiple entries
func sendJsonHWInvByFRUsRsp(w http.ResponseWriter, hf []*sm.HWInvByFRU) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if hf != nil {
		err := json.NewEncoder(w).Encode(hf)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// System Hardware Inventory -  Ordered collection of system hardware.
func sendJsonSystemHWInvRsp(w http.ResponseWriter, hw *sm.SystemHWInventory) {
	http_code := 200
	if hw == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if hw != nil {
		err := json.NewEncoder(w).Encode(hw)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Hardware inventory history of a single component(locational xname or FRU ID).
func sendJsonHWInvHistRsp(w http.ResponseWriter, hh *sm.HWInvHistArray) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if hh != nil {
		err := json.NewEncoder(w).Encode(hh)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Hardware inventory history ordered by locational xname or FRU ID.
func sendJsonHWInvHistArrayRsp(w http.ResponseWriter, hhs *sm.HWInvHistResp) {
	http_code := 200
	if hhs == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if hhs != nil {
		err := json.NewEncoder(w).Encode(hhs)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Individual RedfishEndpoint response, matching a single xname ID.
func sendJsonRFEndpointRsp(w http.ResponseWriter, ep *sm.RedfishEndpoint) {
	http_code := 200
	if ep == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if ep != nil {
		err := json.NewEncoder(w).Encode(ep)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// RedfishEndpoint named array response, e.g. GET on RedfishEndpoints
// collection rather than an individual named entry.
func sendJsonRFEndpointArrayRsp(w http.ResponseWriter, eps *sm.RedfishEndpointArray) {
	http_code := 200
	if eps == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if eps != nil {
		err := json.NewEncoder(w).Encode(eps)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Individual ComponentEndpoint response, matching a single xname ID.  These
// are subcomponents of a RedfishEndpoint created during discovery.
func sendJsonCompEndpointRsp(w http.ResponseWriter, cep *sm.ComponentEndpoint) {
	http_code := 200
	if cep == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if cep != nil {
		err := json.NewEncoder(w).Encode(cep)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// ComponentEndpoint named array response, e.g. GET on RedfishEndpoints
// collection rather than an individual named entry.
func sendJsonCompEndpointArrayRsp(w http.ResponseWriter, ceps *sm.ComponentEndpointArray) {
	http_code := 200
	if ceps == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if ceps != nil {
		err := json.NewEncoder(w).Encode(ceps)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Individual ServiceEndpoint response, matching a single service type and
// xname ID.  These are subcomponents of a RedfishEndpoint created during
// discovery.
func sendJsonServiceEndpointRsp(w http.ResponseWriter, sep *sm.ServiceEndpoint) {
	http_code := 200
	if sep == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if sep != nil {
		err := json.NewEncoder(w).Encode(sep)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// ServiceEndpoint named array response, e.g. GET on ServiceEndpoints
// collection rather than an individual named entry.
func sendJsonServiceEndpointArrayRsp(w http.ResponseWriter, seps *sm.ServiceEndpointArray) {
	http_code := 200
	if seps == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if seps != nil {
		err := json.NewEncoder(w).Encode(seps)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Individual CompEthInterface response, matching a single ID.
func sendJsonCompEthInterfaceRsp(w http.ResponseWriter, cei *sm.CompEthInterface) {
	http_code := 200
	if cei == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if cei != nil {
		err := json.NewEncoder(w).Encode(cei)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of component ethernet interfaces
func sendJsonCompEthInterfaceArrayRsp(w http.ResponseWriter, ceis []*sm.CompEthInterface) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	err := json.NewEncoder(w).Encode(ceis)
	if err != nil {
		fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
	}
}

// Individual CompEthInterface V2 response, matching a single ID.
func sendJsonCompEthInterfaceV2Rsp(w http.ResponseWriter, cei *sm.CompEthInterfaceV2) {
	http_code := 200
	if cei == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if cei != nil {
		err := json.NewEncoder(w).Encode(cei)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of component ethernet interfaces V2
func sendJsonCompEthInterfaceV2ArrayRsp(w http.ResponseWriter, ceis []*sm.CompEthInterfaceV2) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	err := json.NewEncoder(w).Encode(ceis)
	if err != nil {
		fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
	}
}

// Individual CompEthInterface response, matching a single ID.
func sendJsonCompEthInterfaceIPAddressMappingsRsp(w http.ResponseWriter, ipm *sm.IPAddressMapping) {
	http_code := 200
	if ipm == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if ipm != nil {
		err := json.NewEncoder(w).Encode(ipm)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of IPAddress mappings for a component ethernet interface
func sendJsonCompEthInterfaceIPAddressMappingsArrayRsp(w http.ResponseWriter, ceis []sm.IPAddressMapping) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	err := json.NewEncoder(w).Encode(ceis)
	if err != nil {
		fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
	}
}

// Single DiscoveryStatus
func sendJsonDiscoveryStatusRsp(w http.ResponseWriter, stat *sm.DiscoveryStatus) {
	http_code := 200
	if stat == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if stat != nil {
		err := json.NewEncoder(w).Encode(stat)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of DiscoveryStatus
func sendJsonDiscoveryStatusArrayRsp(w http.ResponseWriter, stats []*sm.DiscoveryStatus) {
	http_code := 200
	if len(stats) == 0 {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if len(stats) != 0 {
		err := json.NewEncoder(w).Encode(stats)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of SCN subscriptions
func sendJsonSCNSubscriptionArrayRsp(w http.ResponseWriter, subs *sm.SCNSubscriptionArray) {
	http_code := 200
	if subs == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if subs != nil {
		err := json.NewEncoder(w).Encode(subs)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// A SCN subscription
func sendJsonSCNSubscriptionRsp(w http.ResponseWriter, sub *sm.SCNSubscription) {
	http_code := 200
	if sub == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if sub != nil {
		err := json.NewEncoder(w).Encode(sub)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of groups
func sendJsonGroupArrayRsp(w http.ResponseWriter, groups *[]sm.Group) {
	http_code := 200
	if groups == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if groups != nil {
		err := json.NewEncoder(w).Encode(groups)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// A group
func sendJsonGroupRsp(w http.ResponseWriter, group *sm.Group) {
	http_code := 200
	if group == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if group != nil {
		err := json.NewEncoder(w).Encode(group)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of strings
func sendJsonStringArrayRsp(w http.ResponseWriter, strs *[]string) {
	http_code := 200
	if strs == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if strs != nil {
		err := json.NewEncoder(w).Encode(strs)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of group members
func sendJsonMembersRsp(w http.ResponseWriter, members *sm.Members) {
	http_code := 200
	if members == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if members != nil {
		err := json.NewEncoder(w).Encode(members)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of partitions
func sendJsonPartitionArrayRsp(w http.ResponseWriter, parts *[]sm.Partition) {
	http_code := 200
	if parts == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if parts != nil {
		err := json.NewEncoder(w).Encode(parts)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// A partition
func sendJsonPartitionRsp(w http.ResponseWriter, part *sm.Partition) {
	http_code := 200
	if part == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if part != nil {
		err := json.NewEncoder(w).Encode(part)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of memberships
func sendJsonMembershipArrayRsp(w http.ResponseWriter, memberships []*sm.Membership) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	err := json.NewEncoder(w).Encode(memberships)
	if err != nil {
		fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
	}
}

// A membership
func sendJsonMembershipRsp(w http.ResponseWriter, membership *sm.Membership) {
	http_code := 200
	if membership == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if membership != nil {
		err := json.NewEncoder(w).Encode(membership)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of component locks
func sendJsonCompLockArrayRsp(w http.ResponseWriter, cls []*sm.CompLock) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	err := json.NewEncoder(w).Encode(cls)
	if err != nil {
		fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
	}
}

// A component lock
func sendJsonCompLockRsp(w http.ResponseWriter, cl *sm.CompLock) {
	http_code := 200
	if cl == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if cl != nil {
		err := json.NewEncoder(w).Encode(cl)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Array of v2 component locks
func sendJsonCompLockV2Rsp(w http.ResponseWriter, cls sm.CompLockV2Status) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	err := json.NewEncoder(w).Encode(cls)
	if err != nil {
		fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
	}
}

// Array of v2 component lock update results
func sendJsonCompLockV2UpdateRsp(w http.ResponseWriter, clu sm.CompLockV2UpdateResult) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	err := json.NewEncoder(w).Encode(clu)
	if err != nil {
		fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
	}
}

// Array of v2 component reservations
func sendJsonCompReservationRsp(w http.ResponseWriter, crs sm.CompLockV2ReservationResult) {
	http_code := 200
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	err := json.NewEncoder(w).Encode(crs)
	if err != nil {
		fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
	}
}

// Power Mapping response, single entry
func sendJsonPowerMapRsp(w http.ResponseWriter, m *sm.PowerMap) {
	http_code := 200
	if m == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if m != nil {
		err := json.NewEncoder(w).Encode(m)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// Named PowerMaps array response (i.e. GET on PowerMaps collection)
func sendJsonPowerMapArrayRsp(
	w http.ResponseWriter,
	ms []*sm.PowerMap) {

	http_code := 200
	if ms == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if ms != nil {
		err := json.NewEncoder(w).Encode(ms)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}

// A list of HMS enum values
func sendJsonValueRsp(w http.ResponseWriter, vals *HMSValues) {
	http_code := 200
	if vals == nil {
		http_code = 204
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http_code)
	if vals != nil {
		err := json.NewEncoder(w).Encode(vals)
		if err != nil {
			fmt.Printf("Couldn't encode a JSON command response: %s\n", err)
		}
	}
}
