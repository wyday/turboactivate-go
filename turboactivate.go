// Copyright 2017 wyDay, LLC. All rights reserved.

package turboactivate // import "golang.wyday.com/turboactivate"

/*
#cgo CFLAGS: -I .
#cgo LDFLAGS: -L . -L .. -lTurboActivate

#include "TurboActivate.h"
*/
import "C"
import (
	"errors"
	"strconv"
	"unsafe"
)

// The TurboActivate object.
type TurboActivate struct {
	handle C.uint32_t
}

// IsGenuineResult is the result from the IsGenuine() and IsGenuinEx() functions
type IsGenuineResult int

var (
	// IGRGenuine means the app is activated and genuine
	IGRGenuine IsGenuineResult // Genuine

	// IGRGenuineFeaturesChanged means the app is activated and genuine and the features have changed
	IGRGenuineFeaturesChanged IsGenuineResult = 1 // GenuineFeaturesChanged

	// IGRNotGenuine means the app is not genuine (note: use this in tandem with NotGenuineInVM)
	IGRNotGenuine IsGenuineResult = 2 // NotGenuine

	// IGRNotGenuineInVM means the app is not genuine because you're in a Virtual Machine
	IGRNotGenuineInVM IsGenuineResult = 3 // NotGenuineInVM

	// IGRInternetError should be treated as a warning. That is, tell the user that the activation couldn't be validated with the servers and that they can manually recheck with the servers immediately.
	IGRInternetError IsGenuineResult = 4 // InternetError
)

// TAFlags is the set of flag values you can pass to CheckAndSavePKey(), UseTrial()
type TAFlags int

var (
	// TASystem flag tells TurboActivate to save the activation or
	// trial data on a system-wide basis. This ensures that only
	// a single use has to activate the software on the machine
	// and it will be available to all accounts on the machine.
	TASystem TAFlags = 1

	// TAUser flag tells TurboActivate to save the activation or
	// trial data on a per-user basis.
	TAUser TAFlags = 2

	// TADisallowVM flag is for UseTrial() to disallow trials in virtual machines.
	// If you use this flag in UseTrial() and the customer's machine is a Virtual
	// Machine, then UseTrial() will return an error describing that it's in a VM.
	TADisallowVM TAFlags = 4

	// TAUnverifiedTrial flag is for UseTrial() to tell TurboActivate to use client-side
	// unverified trials. For more information about verified vs. unverified trials,
	// see here: https://wyday.com/limelm/help/trials/
	// Note: unverified trials are unsecured and can be reset by malicious customers.
	TAUnverifiedTrial TAFlags = 16

	// TAVerifiedTrial flag is for UseTrial() to tell it to use verified trials instead
	// of unverified trials. This means the trial is locked to a particular computer.
	// The customer can't reset the trial.
	TAVerifiedTrial TAFlags = 32
)

// TADateCheckFlags is the set of flag values you can pass to IsDateValid()
type TADateCheckFlags int

var (
	// TAHasNotExpired when passed into IsDateValid() verifies
	// that the passed in UTC date-time has not elapsed.
	TAHasNotExpired TADateCheckFlags = 1
)

func taHresultToErr(ret C.HRESULT, funcName string) error {
	switch ret {
	case 0x01: // TA_FAIL
		return errors.New(funcName + " general failure")
	case 0x02: // TA_E_PKEY
		return errors.New("The product key is invalid or there's no product key")
	case 0x03: // TA_E_ACTIVATE
		return errors.New("The product needs to be activated")
	case 0x04: // TA_E_INET
		// More information here: https://wyday.com/limelm/help/faq/#internet-error
		return errors.New("Connection to the servers failed")
	case 0x05: // TA_E_INUSE
		return errors.New("The product key has already been activated with the maximum number of computers")
	case 0x06: // TA_E_REVOKED
		return errors.New("The product key has been revoked")
	case 0x09: // TA_E_TRIAL
		return errors.New("The trial data has been corrupted, using the oldest date possible")
	case 0x0B: // TA_E_COM
		return errors.New("CoInitializeEx failed. Re-enable Windows Management Instrumentation (WMI) service. Contact your system admin for more information")
	case 0x0C: // TA_E_TRIAL_EUSED
		return errors.New("The trial extension has already been used")
	case 0x0D: // TA_E_EXPIRED
		return errors.New("The activation has expired or the system time has been tampered with. Ensure your time, timezone, and date settings are correct. After fixing them restart your computer")
	case 0x0F: // TA_E_PERMISSION
		return errors.New("Insufficient system permission. Either start your process as an admin / elevated user or call the function again with the TA_USER flag")
	case 0x10: // TA_E_INVALID_FLAGS
		return errors.New("The flags you passed to the function were invalid (or missing). Flags like \"TA_SYSTEM\" and \"TA_USER\" are mutually exclusive -- you can only use one or the other")
	case 0x11: // TA_E_IN_VM
		return errors.New("The function failed because this instance of your program is running inside a virtual machine / hypervisor and you've prevented the function from running inside a VM")
	case 0x12: // TA_E_EDATA_LONG
		return errors.New("The \"extra data\" was too long. You're limited to 255 UTF-8 characters. Or, on Windows, a Unicode string that will convert into 255 UTF-8 characters or less")
	case 0x13: // TA_E_INVALID_ARGS
		return errors.New("The arguments passed to the function are invalid. Double check your logic")
	case 0x14: // TA_E_KEY_FOR_TURBOFLOAT
		return errors.New("The product key used is for TurboFloat Server, not TurboActivate")
	case 0x18: // TA_E_NO_MORE_DEACTIVATIONS
		return errors.New("No more deactivations are allowed for the product key. This product is still activated on this computer")
	case 0x19: // TA_E_ACCOUNT_CANCELED
		return errors.New("Can't activate because the LimeLM account is cancelled")
	case 0x1A: // TA_E_ALREADY_ACTIVATED
		return errors.New("You can't use a product key because your app is already activated with a product key. To use a new product key, then first deactivate using either the Deactivate() or DeactivationRequestToFile()")
	case 0x1B: // TA_E_INVALID_HANDLE
		return errors.New("The handle is not valid. You must set a valid VersionGUID when constructing TurboActivate object")
	case 0x1C: // TA_E_ENABLE_NETWORK_ADAPTERS
		// More information here: https://wyday.com/limelm/help/faq/#disabled-adapters
		return errors.New("There are network adapters on the system that are disabled and TurboActivate couldn't read their hardware properties (even after trying and failing to enable the adapters automatically). Enable the network adapters, re-run the function, and TurboActivate will be able to \"remember\" the adapters even if the adapters are disabled in the future")
	case 0x1D: // TA_E_ALREADY_VERIFIED_TRIAL
		return errors.New("The trial is already a verified trial. You need to use the \"TA_VERIFIED_TRIAL\" flag. Can't \"downgrade\" a verified trial to an unverified trial")
	case 0x1E: // TA_E_TRIAL_EXPIRED
		return errors.New("The verified trial has expired. You must request a trial extension from the company")
	case 0x1F: // TA_E_MUST_SPECIFY_TRIAL_TYPE
		return errors.New("You must specify the trial type (TA_UNVERIFIED_TRIAL or TA_VERIFIED_TRIAL). And you can't use both flags. Choose one or the other. We recommend TA_VERIFIED_TRIAL")
	case 0x20: // TA_E_MUST_USE_TRIAL
		return errors.New("You must call TA_UseTrial() before you can get the number of trial days remaining")
	case 0x21: // TA_E_NO_MORE_TRIALS_ALLOWED
		return errors.New("In the LimeLM account either the trial days is set to 0, OR the account is set to not auto-upgrade and thus no more verified trials can be made")
	case 0x22: // TA_E_BROKEN_WMI
		return errors.New("The WMI repository on the computer is broken. To fix the WMI repository see the instructions here: https://wyday.com/limelm/help/faq/#fix-broken-wmi")
	case 0x23: // TA_E_INET_TIMEOUT
		return errors.New("The connection to the server timed out because a long period of time elapsed since the last data was sent or received")
	case 0x24: // TA_E_INET_TLS
		// More information here: https://wyday.com/limelm/help/faq/#internet-error
		return errors.New("The secure connection to the activation servers failed due to a TLS or certificate error. More information here: https://wyday.com/limelm/help/faq/#internet-error")
	default:
		// Make sure you're using the latest turboactivate.go, we occassionally add new error codes
		// and you need latest version of this file to get a detailed description of the error.

		// More information about upgrading here: https://wyday.com/limelm/help/faq/#update-libs

		// You can also view error directly from the source: TurboActivate.h
		return errors.New(funcName + " failed with an unknown error code: " + strconv.FormatUint(uint64(ret), 10))
	}
}

// NewTurboActivate creates a new TurboActivate instance for the provided GUID
func NewTurboActivate(taGUID string, pdetsFilename string) (TurboActivate, error) {

	// Load the TurboActivate.dat file if a path was passed in.
	if pdetsFilename != "" {
		var nativeFilename = getTAStrPtr(pdetsFilename)

		var ret = C.TA_PDetsFromPath(nativeFilename)

		C.free(unsafe.Pointer(nativeFilename))

		// ret != TA_OK && ret != TA_FAIL
		if ret != 0x00 && ret != 0x01 {
			return TurboActivate{}, errors.New("The TurboActivate.dat file failed to load")
		}
	}

	var nativeTaGUID = getTAStrPtr(taGUID)

	var handl C.uint32_t = C.TA_GetHandle(nativeTaGUID)

	C.free(unsafe.Pointer(nativeTaGUID))

	return TurboActivate{
		handle: handl,
	}, nil
}

// Activate activates the product on this computer. You must call "CheckAndSavePKey()
// with a valid product key or have used the TurboActivate Wizard sometime before
// calling this function.
// extraData: Extra data to pass to the LimeLM servers that will be visible for you to see and use.
//            Maximum size is 255 UTF-8 characters
// Returns nil on no error.
func (ta *TurboActivate) Activate(extraData string) error {

	var ret C.HRESULT

	if extraData == "" {
		ret = C.TA_Activate(ta.handle, (*C.ACTIVATE_OPTIONS)(nil))
	} else {

		var actOptions C.ACTIVATE_OPTIONS
		actOptions.nLength = C.uint32_t(unsafe.Sizeof(actOptions))
		actOptions.sExtraData = (C.STRTYPE)(getTAStrPtr(extraData))

		ret = C.TA_Activate(ta.handle, (*C.ACTIVATE_OPTIONS)(unsafe.Pointer(&actOptions)))

		C.free(unsafe.Pointer(actOptions.sExtraData))
	}

	// TA_OK
	if ret == 0x00 {
		return nil
	}

	return taHresultToErr(ret, "Activate")
}

// ActivationRequestToFile gets the "activation request" file for offline activation.
// You must call CheckAndSavePKey() with a valid product key or have used the
// TurboActivate wizard sometime before calling this function.
func (ta *TurboActivate) ActivationRequestToFile(filename string, extraData string) error {

	var nativeFilename = getTAStrPtr(filename)

	var ret C.HRESULT

	if extraData == "" {
		ret = C.TA_ActivationRequestToFile(ta.handle, nativeFilename, (*C.ACTIVATE_OPTIONS)(nil))
	} else {

		var actOptions C.ACTIVATE_OPTIONS

		actOptions.nLength = C.uint32_t(unsafe.Sizeof(actOptions))
		actOptions.sExtraData = (C.STRTYPE)(getTAStrPtr(extraData))

		ret = C.TA_ActivationRequestToFile(ta.handle, nativeFilename, (*C.ACTIVATE_OPTIONS)(unsafe.Pointer(&actOptions)))

		C.free(unsafe.Pointer(actOptions.sExtraData))
	}

	C.free(unsafe.Pointer(nativeFilename))

	// TA_OK
	if ret == 0x00 {
		return nil
	}

	return taHresultToErr(ret, "ActivationRequestToFile")
}

// ActivateFromFile activates from the "activation response" file
// for offline activations.
func (ta *TurboActivate) ActivateFromFile(filename string) error {

	var nativeFilename = getTAStrPtr(filename)

	var ret C.HRESULT = C.TA_ActivateFromFile(ta.handle, nativeFilename)

	C.free(unsafe.Pointer(nativeFilename))

	// TA_OK
	if ret == 0x00 {
		return nil
	}

	return taHresultToErr(ret, "ActivateFromFile")
}

// CheckAndSavePKey verifies (locally) that the product key and saves it if it is.
// this function does not contact the activation servers. Use Activate() to lock the
// product key to a particular machine.
func (ta *TurboActivate) CheckAndSavePKey(productKey string, flags TAFlags) (bool, error) {

	var nativeProductKey = getTAStrPtr(productKey)

	var ret C.HRESULT = C.TA_CheckAndSavePKey(ta.handle, nativeProductKey, C.uint32_t(flags))

	C.free(unsafe.Pointer(nativeProductKey))

	switch ret {
	case 0x00: // TA_OK
		return true, nil

	case 0x01: // TA_FAIL
		return false, nil

	default:
		return false, taHresultToErr(ret, "CheckAndSavePKey")
	}
}

// Deactivate deactivates the product on this computer.
func (ta *TurboActivate) Deactivate(eraseProductKey bool) error {

	var erasePK C.char

	if eraseProductKey {
		erasePK = 1
	} else {
		erasePK = 0
	}

	var ret C.HRESULT = C.TA_Deactivate(ta.handle, erasePK)

	// TA_OK
	if ret == 0x00 {
		return nil
	}

	return taHresultToErr(ret, "Deactivate")
}

// DeactivationRequestToFile get the "deactivation request" file for offline deactivation.
func (ta *TurboActivate) DeactivationRequestToFile(filename string, eraseProductKey bool) error {
	var erasePK C.char

	if eraseProductKey {
		erasePK = 1
	} else {
		erasePK = 0
	}

	var nativeFilename = getTAStrPtr(filename)

	var ret C.HRESULT = C.TA_DeactivationRequestToFile(ta.handle, nativeFilename, erasePK)

	C.free(unsafe.Pointer(nativeFilename))

	// TA_OK
	if ret == 0x00 {
		return nil
	}

	return taHresultToErr(ret, "DeactivationRequestToFile")
}

// GetExtraData gets the extra data value you passed in when activating.
// Returns the extra data if it exists, otherwise it returns an empty string.
func (ta *TurboActivate) GetExtraData() (string, error) {

	var ret C.HRESULT = C.TA_GetExtraData(ta.handle, nil, 0)

	var nativeExtraData = getTAStrBufferPtr(C.size_t(ret))

	ret = C.TA_GetExtraData(ta.handle, nativeExtraData, C.int(ret))

	// TA_OK
	if ret == 0x00 {
		var featureValue = stringFromTAStrPtr(nativeExtraData)

		C.free(unsafe.Pointer(nativeExtraData))

		return featureValue, nil
	}

	C.free(unsafe.Pointer(nativeExtraData))

	return "", taHresultToErr(ret, "GetExtraData")
}

// GetFeatureValue gets the value of a custom license field.
// More information on custom license fields: https://wyday.com/limelm/help/license-features/
func (ta *TurboActivate) GetFeatureValue(featureName string) (string, error) {

	var nativeFeatureName = getTAStrPtr(featureName)

	var ret C.HRESULT = C.TA_GetFeatureValue(ta.handle, nativeFeatureName, nil, 0)

	var nativeFeatureValue = getTAStrBufferPtr(C.size_t(ret))

	ret = C.TA_GetFeatureValue(ta.handle, nativeFeatureName, nativeFeatureValue, C.int(ret))

	C.free(unsafe.Pointer(nativeFeatureName))

	// TA_OK
	if ret == 0x00 {
		var featureValue = stringFromTAStrPtr(nativeFeatureValue)

		C.free(unsafe.Pointer(nativeFeatureValue))

		return featureValue, nil
	}

	C.free(unsafe.Pointer(nativeFeatureValue))

	return "", taHresultToErr(ret, "GetFeatureValue")
}

// GetPKey gets the stored product key. NOTE: if you want to check if a product
// key is valid simply call IsProductKeyValid(). If you want to check if your app
// is locked to the computer then call IsGenuineEx() or IsActivated().
func (ta *TurboActivate) GetPKey() (string, error) {
	var nativePkey = getTAStrBufferPtr(35)

	var ret C.HRESULT = C.TA_GetPKey(ta.handle, nativePkey, 35)

	// TA_OK
	if ret == 0x00 {
		var pkey = stringFromTAStrPtr(nativePkey)

		C.free(unsafe.Pointer(nativePkey))

		return pkey, nil
	}

	C.free(unsafe.Pointer(nativePkey))

	return "", taHresultToErr(ret, "GetPKey")
}

// IsActivated checks whether the computer has been activated.
// Returns true if the computer is activated, false otherwise.
func (ta *TurboActivate) IsActivated() (bool, error) {
	var ret C.HRESULT = C.TA_IsActivated(ta.handle)

	switch ret {
	case 0x00: // TA_OK
		return true, nil

	case 0x01: // TA_FAIL
		return false, nil

	default:
		return false, taHresultToErr(ret, "IsActivated")
	}
}

// IsDateValid checks if the string in the form "YYYY-MM-DD HH:mm:ss" is a valid
// date/time. The date must be in UTC time and "24-hour" format. If your date is
// in some other time format first convert it to UTC time before passing it into
// this function.
// Returns true if the date is valid, false if it's not.
func (ta *TurboActivate) IsDateValid(dateTime string, flags TADateCheckFlags) (bool, error) {

	//TODO: implement
	var nativeDateTime = getTAStrPtr(dateTime)

	var ret C.HRESULT = C.TA_IsDateValid(ta.handle, nativeDateTime, C.uint32_t(flags))

	C.free(unsafe.Pointer(nativeDateTime))

	switch ret {
	case 0x00: // TA_OK
		return true, nil

	case 0x01: // TA_FAIL
		return false, nil

	default:
		return false, taHresultToErr(ret, "IsDateValid")
	}
}

// IsGenuine checks whether the computer is genuinely activated by verifying with the
// LimeLM servers immediately.
// Returns an IsGenuineResult value.
func (ta *TurboActivate) IsGenuine() (IsGenuineResult, error) {

	var ret C.HRESULT = C.TA_IsGenuine(ta.handle)

	switch ret {
	case 0x00: // TA_OK
		return IGRGenuine, nil

	case 0x16: // TA_E_FEATURES_CHANGED
		return IGRGenuineFeaturesChanged, nil

	case 0x04: // TA_E_INET
		return IGRInternetError, nil

	case 0x01, 0x03, 0x06: // TA_FAIL, TA_E_ACTIVATE, TA_E_REVOKED
		return IGRNotGenuine, nil

	case 0x11: // TA_E_IN_VM
		return IGRNotGenuineInVM, nil

	default:
		return IGRNotGenuine, taHresultToErr(ret, "IsGenuine")
	}
}

// IsGenuineEx checks whether the computer is activated, and every "daysBetweenChecks"
// days it check if the customer is genuinely activated by verifying with the
// LimeLM servers.
// daysBetweenChecks: How often to contact the LimeLM servers for validation.
//                    90 days recommended
// graceDaysOnInetErr: If the call fails because of an internet error, how long, in days,
//                     should the grace period last (before returning deactivating and
//                     returning IGRNotGenuine).
//
func (ta *TurboActivate) IsGenuineEx(daysBetweenChecks uint32, graceDaysOnInetErr uint32, skipOffline bool, offlineShowInetErr bool) (IsGenuineResult, error) {

	var gen_opts C.GENUINE_OPTIONS

	gen_opts.nLength = C.uint32_t(unsafe.Sizeof(gen_opts))

	if skipOffline {
		// TA_SKIP_OFFLINE
		gen_opts.flags = 1

		if offlineShowInetErr {
			// TA_OFFLINE_SHOW_INET_ERR
			gen_opts.flags |= 2
		}
	} else {
		gen_opts.flags = 0
	}

	gen_opts.nDaysBetweenChecks = C.uint32_t(daysBetweenChecks)
	gen_opts.nGraceDaysOnInetErr = C.uint32_t(graceDaysOnInetErr)

	var ret C.HRESULT = C.TA_IsGenuineEx(ta.handle, (*C.GENUINE_OPTIONS)(unsafe.Pointer(&gen_opts)))

	switch ret {
	case 0x00: // TA_OK
		return IGRGenuine, nil

	case 0x16: // TA_E_FEATURES_CHANGED
		return IGRGenuineFeaturesChanged, nil

	case 0x04, 0x15: // TA_E_INET, TA_E_INET_DELAYED
		return IGRInternetError, nil

	case 0x01, 0x03, 0x06: // TA_FAIL, TA_E_ACTIVATE, TA_E_REVOKED
		return IGRNotGenuine, nil

	case 0x11: // TA_E_IN_VM
		return IGRNotGenuineInVM, nil

	default:
		return IGRNotGenuine, taHresultToErr(ret, "IsGenuineEx")
	}
}

// GenuineDays gets the number of days until the next time that the IsGenuineEx()
// function contacts the LimeLM activation servers to reverify the activation.
// Returns the number of days remaining and whether the user is in the grace period.
func (ta *TurboActivate) GenuineDays(daysBetweenChecks uint32, graceDaysOnInetErr uint32) (uint32, bool, error) {

	var daysRemain C.uint32_t
	var inGrace C.char

	var ret C.HRESULT = C.TA_GenuineDays(ta.handle, C.uint32_t(daysBetweenChecks), C.uint32_t(graceDaysOnInetErr), (*C.uint32_t)(unsafe.Pointer(&daysRemain)), (*C.char)(unsafe.Pointer(&inGrace)))

	// if != TA_OK
	if ret != 0x00 {
		return 0, false, taHresultToErr(ret, "GenuineDays")
	}

	var inGracePeriod bool = inGrace == 1

	return uint32(daysRemain), inGracePeriod, nil
}

// IsProductKeyValid checks if the product key installed for this product is valid.
// This does NOT check if the product key is activated or genuine.
// Use IsActivated()and IsGenuineEx() instead.
// Returns true if there's a product key that's been saved. False otherwise.
func (ta *TurboActivate) IsProductKeyValid() (bool, error) {
	var ret C.HRESULT = C.TA_IsProductKeyValid(ta.handle)

	switch ret {
	case 0x00: // TA_OK
		return true, nil

	case 0x01: // TA_FAIL
		return false, nil

	default:
		return false, taHresultToErr(ret, "IsProductKeyValid")
	}
}

// SetCustomProxy sets the custom proxy to be used by functions that
// connect to the internet.
func (ta *TurboActivate) SetCustomProxy(proxy string) error {

	var nativeProxy = getTAStrPtr(proxy)

	var ret C.HRESULT = C.TA_SetCustomProxy(nativeProxy)

	C.free(unsafe.Pointer(nativeProxy))

	// TA_OK
	if ret == 0x00 {
		return nil
	}

	return taHresultToErr(ret, "SetCustomProxy")
}

// TrialDaysRemaining gets the number of trial days remaining. You must call
// UseTrial() at least once in the past before calling this function.
// Returns the number of days remaining. 0 days if the trial has expired. (E.g. 1 day means *at most* 1 day. That is it could be 30 seconds.)
func (ta *TurboActivate) TrialDaysRemaining(flags TAFlags) (uint32, error) {

	var daysRemain C.uint32_t

	var ret C.HRESULT = C.TA_TrialDaysRemaining(ta.handle, C.uint32_t(flags), (*C.uint32_t)(unsafe.Pointer(&daysRemain)))

	// TA_OK
	if ret == 0x00 {
		return uint32(daysRemain), nil
	}

	return 0, taHresultToErr(ret, "TrialDaysRemaining")
}

// UseTrial begins the trial the first time it's called.
// Calling it again will validate the trial data hasn't been tampered with.
// Returns true if there was no error and the trial has not expired. Returns false
// if there is no trial or it has already expired or there's an error.
func (ta *TurboActivate) UseTrial(flags TAFlags, extraData string) (bool, error) {

	var ret C.HRESULT

	if extraData == "" {
		ret = C.TA_UseTrial(ta.handle, C.uint32_t(flags), nil)
	} else {
		var nativeExtraData = getTAStrPtr(extraData)

		ret = C.TA_UseTrial(ta.handle, C.uint32_t(flags), nativeExtraData)

		C.free(unsafe.Pointer(nativeExtraData))
	}

	// TA_OK
	if ret == 0x00 {
		return true, nil
	} else if ret == 0x1E { // TA_E_TRIAL_EXPIRED
		return false, nil
	}

	return false, taHresultToErr(ret, "UseTrial")
}

// UseTrialVerifiedRequest generates a "verified trial" offline request file.
// This file will then need to be submitted to LimeLM. You will then need to use
// the TA_UseTrialVerifiedFromFile() function with the response file from LimeLM
// to actually start the trial.
func (ta *TurboActivate) UseTrialVerifiedRequest(filename string, extraData string) error {

	var ret C.HRESULT

	var nativeFilename = getTAStrPtr(filename)

	if extraData == "" {
		ret = C.TA_UseTrialVerifiedRequest(ta.handle, nativeFilename, nil)
	} else {
		var nativeExtraData = getTAStrPtr(extraData)

		ret = C.TA_UseTrialVerifiedRequest(ta.handle, nativeFilename, nativeExtraData)

		C.free(unsafe.Pointer(nativeExtraData))
	}

	C.free(unsafe.Pointer(nativeFilename))

	// TA_OK
	if ret == 0x00 {
		return nil
	}

	return taHresultToErr(ret, "UseTrialVerifiedRequest")
}

// UseTrialVerifiedFromFile uses the "verified trial response" from LimeLM to start the verified trial.
func (ta *TurboActivate) UseTrialVerifiedFromFile(filename string, flags TAFlags) error {

	var ret C.HRESULT

	var nativeFilename = getTAStrPtr(filename)

	ret = C.TA_UseTrialVerifiedFromFile(ta.handle, nativeFilename, C.uint32_t(flags))

	C.free(unsafe.Pointer(nativeFilename))

	// TA_OK
	if ret == 0x00 {
		return nil
	}

	return taHresultToErr(ret, "UseTrialVerifiedFromFile")
}

// ExtendTrial extends the trial using a trial extension created in LimeLM.
func (ta *TurboActivate) ExtendTrial(trialExtension string, flags TAFlags) error {

	var nativeTrialExtension = getTAStrPtr(trialExtension)

	var ret C.HRESULT = C.TA_ExtendTrial(ta.handle, C.uint32_t(flags), nativeTrialExtension)

	C.free(unsafe.Pointer(nativeTrialExtension))

	// TA_OK
	if ret == 0x00 {
		return nil
	}

	return taHresultToErr(ret, "ExtendTrial")
}

// SetCustomActDataPath function allows you to set a custom folder to store the activation
// data files. For normal use we do not recommend you use this function.
//
// Only use this function if you absolutely must store data into a separate
// folder. For example if your application runs on a USB drive and can't write
// any files to the main disk, then you can use this function to save the activation
// data files to a directory on the USB disk.
//
// If you are using this function (which we only recommend for very special use-cases)
// then you must call this function on every start of your program at the very top of
// your app before any other functions are called.
//
// The directory you pass in must already exist. And the process using TurboActivate
// must have permission to create, write, and delete files in that directory.
func (ta *TurboActivate) SetCustomActDataPath(directory string) error {

	var nativeDirectory = getTAStrPtr(directory)

	var ret C.HRESULT = C.TA_SetCustomActDataPath(ta.handle, nativeDirectory)

	C.free(unsafe.Pointer(nativeDirectory))

	// TA_OK
	if ret == 0x00 {
		return nil
	}

	return taHresultToErr(ret, "SetCustomActDataPath")
}
