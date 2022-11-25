package aper

import (
	"fmt"
    //"encoding/hex"
    //"strings"
	//"path"
	"reflect"
	//"runtime"
    //"github.com/davecgh/go-spew/spew"
	//"github.com/free5gc/aper/logger"
)



var ProtocolIEIDMap = map[int64]string{
    0:"AllowedNSSAI",
    1:"AMFName",
    2:"AMFOverloadResponse",
    3:"AMFSetID",
    4:"AMFTNLAssociationFailedToSetupList",
    5:"AMFTNLAssociationSetupList",
    6:"AMFTNLAssociationToAddList",
    7:"AMFTNLAssociationToRemoveList",
    8:"AMFTNLAssociationToUpdateList",
    9:"AMFTrafficLoadReductionIndication",
    10:"AMFUENGAPID",
    11:"AssistanceDataForPaging",
    12:"BroadcastCancelledAreaList",
    13:"BroadcastCompletedAreaList",
    14:"CancelAllWarningMessages",
    15:"Cause",
    16:"CellIDListForRestart",
    17:"ConcurrentWarningMessageInd",
    18:"CoreNetworkAssistanceInformation",
    19:"CriticalityDiagnostics",
    20:"DataCodingScheme",
    21:"DefaultPagingDRX",
    22:"DirectForwardingPathAvailability",
    23:"EmergencyAreaIDListForRestart",
    24:"EmergencyFallbackIndicator",
    25:"EUTRACGI",
    26:"FiveGSTMSI",
    27:"GlobalRANNodeID",
    28:"GUAMI",
    29:"HandoverType",
    30:"IMSVoiceSupportIndicator",
    31:"IndexToRFSP",
    32:"InfoOnRecommendedCellsAndRANNodesForPaging",
    33:"LocationReportingRequestType",
    34:"MaskedIMEISV",
    35:"MessageIdentifier",
    36:"MobilityRestrictionList",
    37:"NASC",
    38:"NASPDU",
    39:"NASSecurityParametersFromNGRAN",
    40:"NewAMFUENGAPID",
    41:"NewSecurityContextInd",
    42:"NGAPMessage",
    43:"NGRANCGI",
    44:"NGRANTraceID",
    45:"NRCGI",
    46:"NRPPaPDU",
    47:"NumberOfBroadcastsRequested",
    48:"OldAMF",
    49:"OverloadStartNSSAIList",
    50:"PagingDRX",
    51:"PagingOrigin",
    52:"PagingPriority",
    53:"PDUSessionResourceAdmittedList",
    54:"PDUSessionResourceFailedToModifyListModRes",
    55:"PDUSessionResourceFailedToSetupListCxtRes",
    56:"PDUSessionResourceFailedToSetupListHOAck",
    57:"PDUSessionResourceFailedToSetupListPSReq",
    58:"PDUSessionResourceFailedToSetupListSURes",
    59:"PDUSessionResourceHandoverList",
    60:"PDUSessionResourceListCxtRelCpl",
    61:"PDUSessionResourceListHORqd",
    62:"PDUSessionResourceModifyListModCfm",
    63:"PDUSessionResourceModifyListModInd",
    64:"PDUSessionResourceModifyListModReq",
    65:"PDUSessionResourceModifyListModRes",
    66:"PDUSessionResourceNotifyList",
    67:"PDUSessionResourceReleasedListNot",
    68:"PDUSessionResourceReleasedListPSAck",
    69:"PDUSessionResourceReleasedListPSFail",
    70:"PDUSessionResourceReleasedListRelRes",
    71:"PDUSessionResourceSetupListCxtReq",
    72:"PDUSessionResourceSetupListCxtRes",
    73:"PDUSessionResourceSetupListHOReq",
    74:"PDUSessionResourceSetupListSUReq",
    75:"PDUSessionResourceSetupListSURes",
    76:"PDUSessionResourceToBeSwitchedDLList",
    77:"PDUSessionResourceSwitchedList",
    78:"PDUSessionResourceToReleaseListHOCmd",
    79:"PDUSessionResourceToReleaseListRelCmd",
    80:"PLMNSupportList",
    81:"PWSFailedCellIDList",
    82:"RANNodeName",
    83:"RANPagingPriority",
    84:"RANStatusTransferTransparentContainer",
    85:"RANUENGAPID",
    86:"RelativeAMFCapacity",
    87:"RepetitionPeriod",
    88:"ResetType",
    89:"RoutingID",
    90:"RRCEstablishmentCause",
    91:"RRCInactiveTransitionReportRequest",
    92:"RRCState",
    93:"SecurityContext",
    94:"SecurityKey",
    95:"SerialNumber",
    96:"ServedGUAMIList",
    97:"SliceSupportList",
    98:"SONConfigurationTransferDL",
    99:"SONConfigurationTransferUL",
    100:"SourceAMFUENGAPID",
    101:"SourceToTargetTransparentContainer",
    102:"SupportedTAList",
    103:"TAIListForPaging",
    104:"TAIListForRestart",
    105:"TargetID",
    106:"TargetToSourceTransparentContainer",
    107:"TimeToWait",
    108:"TraceActivation",
    109:"TraceCollectionEntityIPAddress",
    110:"UEAggregateMaximumBitRate",
    111:"UEAssociatedLogicalNGConnectionList",
    112:"UEContextRequest",
    113:"UENGAPIDs",
    114:"UEPagingIdentity",
    115:"UEPresenceInAreaOfInterestList",
    116:"UERadioCapability",
    117:"UERadioCapabilityForPaging",
    118:"UESecurityCapabilities",
    119:"UnavailableGUAMIList",
    120:"UserLocationInformation",
    121:"WarningAreaList",
    122:"WarningMessageContents",
    123:"WarningSecurityInfo",
    124:"WarningType",
    125:"AdditionalULNGUUPTNLInformation",
    126:"DataForwardingNotPossible",
    127:"DLNGUUPTNLInformation",
    128:"NetworkInstance",
    129:"PDUSessionAggregateMaximumBitRate",
    130:"PDUSessionResourceFailedToModifyListModCfm",
    131:"PDUSessionResourceFailedToSetupListCxtFail",
    132:"PDUSessionResourceListCxtRelReq",
    133:"PDUSessionType",
    134:"QosFlowAddOrModifyRequestList",
    135:"QosFlowSetupRequestList",
    136:"QosFlowToReleaseList",
    137:"SecurityIndication",
    138:"ULNGUUPTNLInformation",
    139:"ULNGUUPTNLModifyList",
    140:"WarningAreaCoordinates",
    141:"PDUSessionResourceSecondaryRATUsageList",
    142:"HandoverFlag",
    143:"SecondaryRATUsageInformation",
    144:"PDUSessionResourceReleaseResponseTransfer",
    145:"RedirectionVoiceFallback",
    146:"UERetentionInformation",
    147:"SNSSAI",
    148:"PSCellInformation",
    149:"LastEUTRANPLMNIdentity",
    150:"MaximumIntegrityProtectedDataRateDL",
    151:"AdditionalDLForwardingUPTNLInformation",
    152:"AdditionalDLUPTNLInformationForHOList",
    153:"AdditionalNGUUPTNLInformation",
    154:"AdditionalDLQosFlowPerTNLInformation",
    155:"SecurityResult",
    156:"ENDCSONConfigurationTransferDL",
    157:"ENDCSONConfigurationTransferUL",
}

var ProcedureCodeMap = map[int64]string{
    0:"AMFConfigurationUpdate",
    1:"AMFStatusIndication",
    2:"CellTrafficTrace",
    3:"DeactivateTrace",
    4:"DownlinkNASTransport",
    5:"DownlinkNonUEAssociatedNRPPaTransport",
    6:"DownlinkRANConfigurationTransfer",
    7:"DownlinkRANStatusTransfer",
    8:"DownlinkUEAssociatedNRPPaTransport",
    9:"ErrorIndication",
    10:"HandoverCancel",
    11:"HandoverNotification",
    12:"HandoverPreparation",
    13:"HandoverResourceAllocation",
    14:"InitialContextSetup",
    15:"InitialUEMessage",
    16:"LocationReportingControl",
    17:"LocationReportingFailureIndication",
    18:"LocationReport",
    19:"NASNonDeliveryIndication",
    20:"NGReset",
    21:"NGSetup",
    22:"OverloadStart",
    23:"OverloadStop",
    24:"Paging",
    25:"PathSwitchRequest",
    26:"PDUSessionResourceModify",
    27:"PDUSessionResourceModifyIndication",
    28:"PDUSessionResourceRelease",
    29:"PDUSessionResourceSetup",
    30:"PDUSessionResourceNotify",
    31:"PrivateMessage",
    32:"PWSCancel",
    33:"PWSFailureIndication",
    34:"PWSRestartIndication",
    35:"RANConfigurationUpdate",
    36:"RerouteNASRequest",
    37:"RRCInactiveTransitionReport",
    38:"TraceFailureIndication",
    39:"TraceStart",
    40:"UEContextModification",
    41:"UEContextRelease",
    42:"UEContextReleaseRequest",
    43:"UERadioCapabilityCheck",
    44:"UERadioCapabilityInfoIndication",
    45:"UETNLABindingRelease",
    46:"UplinkNASTransport",
    47:"UplinkNonUEAssociatedNRPPaTransport",
    48:"UplinkRANConfigurationTransfer",
    49:"UplinkRANStatusTransfer",
    50:"UplinkUEAssociatedNRPPaTransport",
    51:"WriteReplaceWarning",
    52:"SecondaryRATDataUsageReport",
}


type mappingFunc func(reflect.Value, int, string) (reflect.Value, error)


var CustomFieldValues = map[string]mappingFunc{
    "ProtocolIEName":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        protCode := val.Field(fieldIdx-1).Int()
        return reflect.ValueOf(ProtocolIEIDMap[protCode]), nil
    },
    "ProcedureName":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        procCode := val.Field(fieldIdx-1).Int()
        return reflect.ValueOf(ProcedureCodeMap[procCode]), nil
    },
    "PLMNMCC":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var octetStr OctetString = val.FieldByName("Value").Interface().(OctetString)
        b0 := int64(octetStr.Bytes[0])
        b1 := int64(octetStr.Bytes[1])
        mcc := ((b0 & 0x0F) * 100 ) +(b0>>4 *10) + (b1 & 0x0F) 
        return reflect.ValueOf(mcc), nil
    },
    "PLMNMNC":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var octetStr OctetString = val.FieldByName("Value").Interface().(OctetString)
        b1 := int64(octetStr.Bytes[1])
        b2 := int64(octetStr.Bytes[2])
        var mnc int64 = 0
        if b1>>4 != 0x0F {
            mnc += b1>>4 * 100
        }
        mnc += (b2 & 0x0F) * 10
        mnc += b2>>4 
        return reflect.ValueOf(mnc), nil
    },
    "EUTRAEEA1":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf((b0 >> 7) == 1)
        return ret, nil
    },
    "EUTRAEEA2":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf(((b0 & 0x40) >> 6) == 1)
        return ret, nil
    },
    "EUTRAEEA3":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf(((b0 & 0x20) >> 5) == 1)
        return ret, nil
    },
    "EUTRAEIA1":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf((b0 >> 7) == 1)
        return ret, nil
    },
    "EUTRAEIA2":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf(((b0 & 0x40) >> 6) == 1)
        return ret, nil
    },
    "EUTRAEIA3":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf(((b0 & 0x20) >> 5) == 1)
        return ret, nil
    },
    "NRNEA1":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf((b0 >> 7) == 1)
        return ret, nil
    },
    "NRNEA2":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf(((b0 & 0x40) >> 6) == 1)
        return ret, nil
    },
    "NRNEA3":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf(((b0 & 0x20) >> 5) == 1)
        return ret, nil
    },
    "NRNIA1":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf((b0 >> 7) == 1)
        return ret, nil
    },
    "NRNIA2":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf(((b0 & 0x40) >> 6) == 1)
        return ret, nil
    },
    "NRNIA3":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        b0 := int64(bitStr.Bytes[0])
        ret := reflect.ValueOf(((b0 & 0x20) >> 5) == 1)
        return ret, nil
    },
    "IPv4":func(val reflect.Value, fieldIdx int, fieldName string) (reflect.Value, error){
        var bitStr BitString = val.FieldByName("Value").Interface().(BitString)
        bytes := bitStr.Bytes
        if len(bytes) != 4{
            return reflect.ValueOf(""), fmt.Errorf("Invalid IP Address: %v", bytes)  
        }
        ipv4 := fmt.Sprintf("%d.%d.%d.%d",bytes[0],bytes[1],bytes[2],bytes[3])
        return reflect.ValueOf(ipv4), nil
    },

    
}

