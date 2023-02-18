package errors

/*
To add new error enum, just add the desire name in the end of the const. No need to add value on it,
It will automatically generate value
*/
type IEIDError int
type GenericError int
type ieidMAP map[IEIDError]string
type gameErrorMAP map[GenericError]string

const (
	VALIDATE_GAME_ERROR GenericError = iota + 0
	VALIDATE_SELECTION_HEADER_STATUS_ERROR
	VALIDATE_SELECTION_LINE_ATTRIBUTE_ERROR
	AUTHENTICATION_FAILED
	VALIDATE_EVENT_ERROR
	VALIDATE_STAGE_ERROR
	VALIDATE_ODDS_ERROR
	VALIDATE_SELECTION_ERROR
	VALIDATE_MARKET_TYPE_ERROR
	VALIDATE_GAME_STATE_ERROR
	LESS_THAN_MIN_BET_AMOUNT
	MORE_THAN_MAX_BET_AMOUNT
	ACCOUNT_IS_FROZEN
	USER_IS_IN_SLEEP_STATUS
	MINI_GAME_DISABLED
	WALLET_ERROR
	CHART_NOT_AVAILABLE_IN_CONFIG
	VALIDATE_GAME_SOUND_ERROR
	VALIDATE_EFFECTS_SOUND_ERROR
	ALREADY_RUNNING_AUTO_PLAY
	EVENT_CLOSED
	VALIDATE_MAX_NUM_ROUND_ERROR
	VALIDATE_ON_WIN_BET_AMOUNT_CHANGE_ERROR
	VALIDATE_ON_LOSS_BET_AMOUNT_CHANGE_ERROR
	VALIDATE_STATUS_ERROR
	VALIDATE_AUTO_PLAY_BET_ERROR
	QUERY_PARAM_ERROR
	VALIDATE_BET_SELECTION_TYPE_ERROR
	ESPORTS_API_FAILED
	BET_TYPE_SYNC_POST_REQUEST_EXCEPTION
	BET_TYPE_SYNC_PATCH_REQUEST_EXCEPTION
	COMPETITION_SYNC_POST_REQUEST_EXCEPTION
	COMPETITION_SYNC_PATCH_REQUEST_EXCEPTION
	EVENT_SYNC_POST_REQUEST_EXCEPTION
	EVENT_SYNC_PATCH_REQUEST_EXCEPTION
	COMPETITION_SYNC_STATUS_EXCEPTION
	GAME_SYNC_POST_REQUEST_EXCEPTION
	GAME_SYNC_PATCH_REQUEST_EXCEPTION
	ES_MODEL_DEPENDENCY_EXCEPTION
	MARKET_SYNC_POST_REQUEST_EXCEPTION
	MARKET_SYNC_PATCH_REQUEST_EXCEPTION
	RESULT_SYNC_POST_REQUEST_EXCEPTION
	RESULT_SYNC_PATCH_REQUEST_EXCEPTION
	TICKET_SYNC_POST_REQUEST_EXCEPTION
	TICKET_SYNC_PATCH_REQUEST_EXCEPTION
	TICKET_INVALID_SETTLEMENT_DATE_EXCEPTION
	NOT_IMPLEMENTED_ERROR
	ENOUGH_FUTURE_EVENT_ERROR
	NO_AVAILABLE_HASH_SEQUENCE_ERROR
	VALIDATE_MARKET_ERROR
	UNSYNCED_MARKET_EXCEPTION
	EMPTY_RESULT_DATA_EXCEPTION
	EVENT_STATUS_ERROR
	NONE_MARKET_EXCEPTION
	GAME_MANAGER_NOT_SET_ERROR
	UNHANDLED_MINI_GAME_ERROR
	UNAVAILABLE_MINI_GAME_TABLE_ERROR
	NOT_ENOUGH_BALANCE_EXCEPTION

	// Channels
	CHANNEL_FULL
	MISSING_ROOM_GROUP_NAME
	DO_NOT_SEND

	//Golang new  err
	MINI_GAME_TABLE_ERROR
	SKIP_LIMIT_REACHED
	NO_RECORD_FOUND
	DUPLICATE_BET
	PATCH_ERROR
	TICKET_CREATION_ERROR
)

const (
	IEID_EVENT_STOP_BETTING_PHASE IEIDError = iota + 1
	IEID_EVENT_ALREADY_PLACED_BET
	IEID_ACCOUNT_IS_FROZEN
	IEID_EVENT_MAX_BET_LIMIT
	IEID_MAX_BET_LIMIT
	IEID_MIN_BET_LIMIT
	IEID_MARKET_TYPE_ERROR
	IEID_EVENT_NOT_FOUND
	IEID_WALLET_ERROR
	IEID_INVALID_SELECTION
	IEID_SKIP_LIMIT_REACHED
	IEID_EVENT_NO_ACTIVE
	IEID_NO_COMBO_BET_FOUND
	IEID_MAX_LOL_TOWER_LIMIT
	IEID_DUPLICATE_BET
	IEID_EVENT_NOT_ON_BETTING_PHASE
	IEID_TICKET_CREATION_ERROR
	IEID_WALLET_ERROR_INIT
	IEID_WALLET_ERROR_ROLLBACK
	IEID_WALLET_ERROR_COMMIT
	IEID_MEMBER_MAX_PAYOUT_LIMIT
	IEID_NO_ACTIVE_TICKET
	IEID_HAS_ACTIVE_MAIN_TICKET
	IEID_PATCH_ERROR
	IEID_BET_PROCESS_ONGOING
	IEID_ESPORTS_API_FAILED
	IEID_NOT_ENOUGH_BALANCE
	IEID_GAME_DISABLED
)

var ERROR_MAP = gameErrorMAP{
	VALIDATE_GAME_ERROR:               "Invalid Game",
	AUTHENTICATION_FAILED:             "Authentication Failed",
	VALIDATE_EVENT_ERROR:              "Invalid Event",
	ACCOUNT_IS_FROZEN:                 "Account is Frozen",
	USER_IS_IN_SLEEP_STATUS:           "User is in sleep status",
	MINI_GAME_DISABLED:                "Table is disabled",
	VALIDATE_ODDS_ERROR:               "Invalid Odds",
	NOT_ENOUGH_BALANCE_EXCEPTION:      "Not enought balance",
	EVENT_STATUS_ERROR:                "Event status error",
	NOT_IMPLEMENTED_ERROR:             "Not implemented",
	LESS_THAN_MIN_BET_AMOUNT:          "Bet amount is less than min bet amount",
	MORE_THAN_MAX_BET_AMOUNT:          "Bet amount is greater than min bet amount",
	WALLET_ERROR:                      "Error processing wallet transaction",
	SKIP_LIMIT_REACHED:                "Skil limit reached",
	UNAVAILABLE_MINI_GAME_TABLE_ERROR: "Unavailable mini game table",
	NO_RECORD_FOUND:                   "No record found",
	DUPLICATE_BET:                     "Bet already exists",
	VALIDATE_BET_SELECTION_TYPE_ERROR: "Invalid Selection",
	PATCH_ERROR:                       "Unable to patch data",
	ESPORTS_API_FAILED:                "Esports API Failed",
}

var ERROR_MAP_IEID = ieidMAP{
	IEID_EVENT_STOP_BETTING_PHASE:   "Event is on stop betting phase",
	IEID_EVENT_ALREADY_PLACED_BET:   "User already placed bet on the event",
	IEID_ACCOUNT_IS_FROZEN:          "Account is Frozen",
	IEID_EVENT_MAX_BET_LIMIT:        "Event Details Max bet limit is reached",
	IEID_MAX_BET_LIMIT:              "Bet Amount is higher than Member Max bet limit",
	IEID_MIN_BET_LIMIT:              "Bet Amount is below Member Min bet limit",
	IEID_MARKET_TYPE_ERROR:          "Invalid market type",
	IEID_EVENT_NOT_FOUND:            "Event not found",
	IEID_WALLET_ERROR:               "Error processing wallet transaction",
	IEID_INVALID_SELECTION:          "Invalid selection",
	IEID_SKIP_LIMIT_REACHED:         "Skip limit reached",
	IEID_EVENT_NO_ACTIVE:            "No active event",
	IEID_NO_COMBO_BET_FOUND:         "No combo bet found",
	IEID_MAX_LOL_TOWER_LIMIT:        "Max tower level has been reached",
	IEID_DUPLICATE_BET:              "Bet already exists",
	IEID_EVENT_NOT_ON_BETTING_PHASE: "Event is not on betting phase",
	IEID_TICKET_CREATION_ERROR:      "Failed to create ticket",
	IEID_WALLET_ERROR_INIT:          "Wallet transaction failed",
	IEID_WALLET_ERROR_ROLLBACK:      "Wallet transaction rollback failed",
	IEID_WALLET_ERROR_COMMIT:        "Wallet transaction commit failed",
	IEID_MEMBER_MAX_PAYOUT_LIMIT:    "Max Payout is reached",
	IEID_NO_ACTIVE_TICKET:           "No Active Bet",
	IEID_HAS_ACTIVE_MAIN_TICKET:     "You still have an active bet",
	IEID_PATCH_ERROR:                "Unable to patch data",
	IEID_BET_PROCESS_ONGOING:        "Ongoing process for bet",
	IEID_ESPORTS_API_FAILED:         "Esports API Failed",
	IEID_NOT_ENOUGH_BALANCE:         "Not Enough Balance",
	IEID_GAME_DISABLED:              "Table is disabled",
}