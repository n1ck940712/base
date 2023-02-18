package service

import (
	"fmt"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"

	"gorm.io/gorm"
)

type TicketService struct {
	Ticket               *models.Ticket
	ComboTicket          *models.ComboTicket
	TicketsForSettlement *models.TicketsForSettlement
}

type LevelSkipCnt struct {
	Level int16
	Skip  int8
}

type currentLevelPayout struct {
	CurrentLevel int8    `gorm:"column:current_level"`
	BetAmount    float64 `gorm:"column:bet_amount"`
	MaxPayout    float64 `gorm:"column:max_payout_amount"`
}

type ITicketService interface {
	GetMemberLevel(UserID int64) LevelSkipCnt
	GetBetDetails(id string) (*models.Ticket, error)
	CreateTicket(ticket *models.Ticket, cTicket *models.ComboTicket, tableID int64, callback func() error) error
	GetMemberActiveTicket(UserID int64) models.Ticket
	GetMemberLatestComboTicket(TicketID string) models.ComboTicket
	UpdateTicket(ticket *models.Ticket, data models.Ticket) error
	UpdateTicketStatus(ticketID string, status int) (*models.Ticket, error)
	CreateComboTicket(data models.ComboTicket) (models.ComboTicket, error)
	GetTableTicketsForSettlement(tableID int64) *[]models.TicketsForSettlement
	UpdateMultipleTickets(ticketIDs string, status int64) bool
	UpdateComboTickets(ticketIDs string, status int64) bool
	CheckTicketIfUnattended(tableID int64, ticket models.TicketsForSettlement) bool
	MemberHasTicketOnEvent(eventID int64, userID int64) bool
	GetMemberTicketPayoutDetails(userID int64, offset int32) currentLevelPayout
	GetOldUnsettledTicketsForSettlement(event models.Event, tableID int64) *[]models.OldTicketsForSettlement
}

func NewTicket() *TicketService {
	return &TicketService{}
}

func (t *TicketService) CreateTicket(ticket *models.Ticket, cTicket *models.ComboTicket, tableID int64, callback func() error) error {
	result := DB.Transaction(func(tx *gorm.DB) error {
		var user models.MemberTable
		if err := tx.Raw(`
		SELECT 
			* 
		FROM mini_game_memberminigametable 
		WHERE 
			mini_game_table_id=? 
			AND user_id=? FOR UPDATE
		`, tableID, ticket.UserID).First(&user).Error; err != nil {
			return err
		}
		time := time.Now()
		ticket.Ctime = time
		ticket.Mtime = time
		//create ticket
		if err := tx.Create(&ticket).Error; err != nil {
			return err
		}
		if cTicket != nil {
			cTicket.Ctime = time
			cTicket.Mtime = time
			//create combo ticket
			if err := tx.Create(&cTicket).Error; err != nil {
				return err
			}
		}
		return callback()
	})

	return result
}

func (t *TicketService) CreateComboTicket(cTicket *models.ComboTicket, tableID int64) error {
	result := DB.Transaction(func(tx *gorm.DB) error {
		var user models.MemberTable
		if err := tx.Raw(`
		SELECT 
			* 
		FROM mini_game_memberminigametable 
		WHERE 
			mini_game_table_id=? 
			AND user_id=? FOR UPDATE`, tableID, cTicket.UserID).First(&user).Error; err != nil {
			return err
		}
		time := time.Now()
		cTicket.Ctime = time
		cTicket.Mtime = time
		result := tx.Create(&cTicket)

		return result.Error
	})

	return result
}

func (t *TicketService) GetMemberActiveTicket(UserID int64) *models.Ticket {
	tt := time.Now().Add(-(constants.LOL_TOWER_GAME_DURATION * 10) * time.Second)
	result := DB.Raw(`
		SELECT mgt.* 
		FROM mini_game_ticket mgt
		JOIN mini_game_event mge on mge.id = mgt.event_id
		WHERE mgt.user_id = ? and mgt.status = ? 
		AND mge.start_datetime >= ?
		ORDER BY mgt.ctime DESC`,
		UserID,
		constants.TICKET_STATUS_PAYMENT_CONFIRMED,
		tt).First(&t.Ticket)

	if result.Error == gorm.ErrRecordNotFound {
		logger.Error("GetMemberActiveTicket Error")
		return nil
	}

	return t.Ticket
}

func (t *TicketService) UpdateTicket(ticket *models.Ticket, data models.Ticket) error {
	return DB.Model(&ticket).Updates(data).Error
}

func (t *TicketService) UpdateTicketStatus(ticketID string, status int) (*models.Ticket, error) {
	temp := map[string]interface{}{"status": status, "mtime": time.Now()}
	result := DB.Table("mini_game_ticket").Where("id = ?", ticketID).Updates(temp)

	if result.Error != nil {
		logger.Error("Update ticket Error")
		return nil, result.Error
	}

	ticket, _ := t.GetBetDetails(ticketID)
	return ticket, nil
}

func (t *TicketService) GetBetDetails(ticketID string) (*models.Ticket, error) {

	result := DB.Raw(`
		SELECT * 
		FROM mini_game_ticket 
		WHERE id =?`,
		ticketID).First(&t.Ticket)

	if result.Error == gorm.ErrRecordNotFound {
		logger.Error("Ticket not found")
		return nil, result.Error
	}

	return t.Ticket, nil
}

func (t *TicketService) GetMemberLatestComboTicket(ticketID string) *models.ComboTicket {
	result := DB.Raw(`
		SELECT * 
		FROM mini_game_combo_ticket 
		WHERE ticket_id = ? ORDER BY ctime DESC`,
		ticketID).First(&t.ComboTicket)

	if result.Error == gorm.ErrRecordNotFound {
		logger.Error("GetMemberActiveComboTicket Error")
		return nil
	}

	return t.ComboTicket
}

func (t *TicketService) GetMemberLevel(param map[string]interface{}) LevelSkipCnt {
	var levelSkipCnt LevelSkipCnt
	result := DB.Raw(`
		SELECT l.level, l.skip from mini_game_ticket t
		JOIN mini_game_lol_tower_member_level l on l.ticket_id = t.id
		WHERE
		mini_game_table_id = ?
		and t.user_id = ?
		ORDER BY l.ctime DESC LIMIT 1`,
		param["table_id"],
		param["user_id"],
	).First(&levelSkipCnt)
	// todo: // ticket status for active
	fmt.Println("get member level ----", levelSkipCnt)
	if result.Error != nil {
		logger.Error("GetMemberLevel Error:")
	}
	return levelSkipCnt
}

func (t *TicketService) GetMemberLOLActiveCombo(param map[string]interface{}) {
	DB.Raw("SELECT * mini_game_ticket t WHERE user_id = ? AND mini_game_table_id = ? ORDER BY ctime DESC",
		param["user_id"], param["table_id"])
}

func (t *TicketService) MemberHasTicketOnEvent(eventID int64, userID int64) bool {
	var num int64
	// ignore status TICKET_STATUS_PAYMENT_FAILED and TICKET_STATUS_CANCELLED
	DB.Table("mini_game_combo_ticket").
		Where("event_id = ?", eventID).
		Where("user_id = ?", userID).
		Not(map[string]interface{}{"status": []int{constants.TICKET_STATUS_PAYMENT_FAILED, constants.TICKET_STATUS_CANCELLED}}).
		Count(&num)
	return num > 0
}

func (t *TicketService) GetTableTicketsForSettlement(tableID int64) *[]models.TicketsForSettlement {
	var tickets []models.TicketsForSettlement
	res := DB.Raw(`
			SELECT ticket_id, max(ct.id), max(ct.ctime) AS combo_ctime, count(*) AS count FROM mini_game_combo_ticket ct 
			JOIN mini_game_event e on e.id = ct.event_id
			WHERE e.status = ?
			AND ct.status = ?
			AND ct.mini_game_table_id = ?
			AND ct.event_id < (
				SELECT id FROM mini_game_event WHERE status = ? AND mini_game_table_id = ? ORDER BY id DESC LIMIT 1
			)
			GROUP BY ticket_id
			ORDER BY combo_ctime DESC
		`,
		constants.EVENT_STATUS_FOR_SETTLEMENT,
		constants.TICKET_STATUS_PAYMENT_CONFIRMED,
		tableID,
		constants.EVENT_STATUS_ACTIVE,
		tableID,
	).Find(&tickets)

	if res.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return &tickets
}

func (t *TicketService) GetEventResult(eventID int64) []string {
	type result struct {
		Value string
	}
	var temp result

	res := DB.Raw(`
		SELECT
			value
		FROM
			mini_game_event e
			JOIN mini_game_eventresult r ON e.ID = r.event_id 
		WHERE
			e.ID = ?
		`, eventID,
	).Find(&temp)

	if res.Error == gorm.ErrRecordNotFound {
		logger.Error("Event result not found")
	}

	eRes := strings.Split(temp.Value, ",")
	return eRes
}

func (t *TicketService) GetMaxPayoutTickets(eventID int64, tableID int64) *[]models.TicketsForSettlement {
	var tickets []models.TicketsForSettlement

	rawQuery := fmt.Sprintf(`
		SELECT
			ct.ticket_id,
			ct.status,
			MAX ( ct.ID ) AS combo_ticket_id,
			COUNT ( 1 ),
			MAX ( event_id ) AS event_id
		FROM mini_game_combo_ticket ct
		JOIN mini_game_event e ON e.ID = ct.event_id
		JOIN mini_game_memberminigametable mgt ON ( mgt.user_id = ct.user_id AND mgt.is_enabled = TRUE AND mgt.mini_game_table_id = ct.mini_game_table_id)
		JOIN mini_game_lol_tower_member_level member_level ON ( member_level.combo_ticket_id = ct.ID ) 
		WHERE
			ct.status = %d
			AND ct.mini_game_table_id = %d
			AND ct.event_id = %d 
			AND ( member_level.next_level_odds * ct.amount ) >= mgt.max_payout_amount 
			AND member_level.next_level_odds IS NOT null 
			AND ct.result = %d
		GROUP BY
		ct.ticket_id,
		ct.status
	`, constants.TICKET_STATUS_PAYMENT_CONFIRMED, tableID, eventID, constants.TICKET_RESULT_WIN)

	DB.Raw(rawQuery).Scan(&tickets)

	return &tickets
}

func (t *TicketService) GetEventLossTickets(eventID int64) *[]models.TicketsForSettlement {
	var tickets []models.TicketsForSettlement
	res := DB.Raw(`
			SELECT ticket_id, max(ct.id), max(ct.ctime) AS combo_ctime, count(*) AS count FROM mini_game_combo_ticket ct
			JOIN mini_game_event e on e.id = ct.event_id
			WHERE ct.status = ?
			AND ct.result = ?
			AND ct.event_id = ?
			GROUP BY ticket_id
			ORDER BY combo_ctime DESC
		`,
		constants.TICKET_STATUS_PAYMENT_CONFIRMED,
		constants.TICKET_RESULT_LOSS,
		eventID,
	).Find(&tickets)

	if res.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return &tickets
}

func (t *TicketService) UpdateMultipleTickets(ticketIDs []string, status int) bool {

	result := DB.Exec(`
			UPDATE mini_game_ticket 
			SET status = ?, mtime = NOW(), status_mtime = NOW(), local_data_version = local_data_version+1
			WHERE id IN (?)
		`,
		status, ticketIDs,
	)

	if result.Error != nil {
		logger.Error("Update ticket Error")
		return false
	}
	return true
}

func (t *TicketService) UpdateComboTickets(ticketIDs []string, status int) bool {
	result := DB.Exec(`
			UPDATE mini_game_combo_ticket 
			SET status = ?, mtime = NOW(), status_mtime = NOW(), local_data_version = local_data_version+1
			WHERE ticket_id IN (?)
		`,
		status, ticketIDs,
	)

	if result.Error != nil {
		logger.Error("Update combo ticket Error")
		return false
	}
	return true
}

func (t *TicketService) GetUnattendedTickets(eventID int64, prevEventID int64, tableID int64) *[]models.TicketsForSettlement {
	var tickets []models.TicketsForSettlement
	res := DB.Raw(`
	SELECT
		*
	FROM
		(
		SELECT
			ticket_id,
			MAX ( ct.ID ) AS combo_ticket_id,
			COUNT ( 1 ),
			MAX ( event_id ) AS event_id
		FROM
			mini_game_combo_ticket ct
			JOIN mini_game_event e ON e.ID = ct.event_id 
		WHERE
			ct.status = ?
			AND ct.mini_game_table_id = ?
			AND ct.event_id <= ? 
			OR (ct.selection = ? AND event_id < ?)
		GROUP BY
			ticket_id
		) comp 
	WHERE
		comp.event_id = ?
		`,
		constants.TICKET_STATUS_PAYMENT_CONFIRMED,
		tableID,
		eventID,
		constants.LOL_SKIP_SELECTION,
		prevEventID,
		prevEventID,
	).Find(&tickets)

	if res.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return &tickets
}

func (t *TicketService) CheckTicketIfUnattended(tableID int64, ticket models.TicketsForSettlement) bool {
	type resStruct struct {
		TotalRows int64 `json:"total_rows"`
	}
	var result resStruct
	res := DB.Raw(`
	SELECT
		count(1) AS total_rows
	FROM
		mini_game_event 
	WHERE
		ID > ? 
		AND mini_game_table_id = ?
		AND status IN ( ?, ?, ? )
	`,
		ticket.EventID,
		tableID,
		constants.EVENT_STATUS_FOR_SETTLEMENT,
		constants.EVENT_STATUS_SETTLEMENT_IN_PROGRESS,
		constants.EVENT_STATUS_SETTLED,
	).Find(&result)

	if res.Error == gorm.ErrRecordNotFound {
		return false
	}

	logger.Info("unattended ahead events---------", ticket.EventID, result)
	return result.TotalRows > 1
}

func (t *TicketService) GetEventMaxLevelTickets(eventID int64, tableID int64) *[]models.TicketsForSettlement {
	var tickets []models.TicketsForSettlement
	res := DB.Raw(`
	SELECT 
		tower_level.ticket_id, 
		tower_level.combo_ticket_id, 
		combo_ticket.status, 
		combo_ticket.event_id, 
		tower_level.level, 
		tower_level.skip
	FROM 
		mini_game_lol_tower_member_level AS tower_level
		LEFT JOIN mini_game_combo_ticket AS combo_ticket ON tower_level.combo_ticket_id = combo_ticket.id
	WHERE 
		combo_ticket.status = ?
		AND combo_ticket.mini_game_table_id = ?
		AND (tower_level.level-tower_level.skip) = ?
		AND combo_ticket.event_id = ?
	`, constants.TICKET_STATUS_PAYMENT_CONFIRMED, tableID, constants.LOL_TOWER_MAX_LEVEL-constants.LOL_TOWER_SKIP_COUNT, eventID).Find(&tickets)

	if res.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return &tickets
}

func (t *TicketService) GetMemberTicketPayoutDetails(userID int64, tableID int64, offset int32) (currentLevelPayout, error) {
	result := currentLevelPayout{}
	rawQuery := fmt.Sprintf(`
	SELECT
		combo_ticket.amount AS bet_amount,
		tower_level.level + 1 AS current_level,
		member_table.max_payout_amount
	FROM 
		mini_game_lol_tower_member_level tower_level
		INNER JOIN mini_game_combo_ticket combo_ticket ON combo_ticket.id = tower_level.combo_ticket_id
		INNER JOIN mini_game_event mg_event ON mg_event.id = combo_ticket.event_id
		INNER JOIN mini_game_memberminigametable member_table ON member_table.user_id = tower_level.user_id
	WHERE 
		tower_level.user_id = %d
		AND combo_ticket.result = 0
		AND combo_ticket.mini_game_table_id = %d
		AND mg_event.start_datetime > (now() - INTERVAL '%d seconds')
		AND member_table.mini_game_table_id = combo_ticket.mini_game_table_id
	ORDER BY tower_level.ctime DESC
	`, userID, tableID, offset)
	dbResult := DB.Raw(rawQuery).First(&result)

	if dbResult.Error != nil {
		return result, dbResult.Error
	}

	return result, nil
}

func (t *TicketService) GetTotalWinlossAmount(startDateTime time.Time, endDateTime time.Time, tableID int64) float64 {
	var res = make([]float64, 0)

	DB.Raw(`
	SELECT 
		CASE WHEN SUM( ticket.win_loss_amount ) IS NULL THEN 0 ELSE SUM ( ticket.win_loss_amount/ticket.exchange_rate ) END AS total_win_loss 
	FROM
		mini_game_ticket ticket
	WHERE
		ticket.ctime BETWEEN ?
		AND ?
		AND ticket.status >= ?
		AND mini_game_table_id = ?
	`, startDateTime, endDateTime, constants.TICKET_STATUS_SETTLEMENT_IN_PROGRESS, tableID).Pluck("total_win_loss", &res)
	return res[0]
}

func (t *TicketService) GetOldUnsettledTicketsForSettlement(event models.Event, tableID int64) *[]models.OldTicketsForSettlement {
	// settle old tickets less than 10 whole games
	prev3EventStartDatetime := time.Now().Add(-(constants.LOL_TOWER_GAME_DURATION * 10) * time.Second)
	var tickets []models.OldTicketsForSettlement
	res := DB.Preload("ComboTickets").Raw(`
		SELECT
			* 
		FROM
			mini_game_ticket t 
		WHERE
			t.ctime < ?
			AND t.status < ?
			AND t.mini_game_table_id = ?
			AND t.event_id NOT IN(?,?)
		`,
		prev3EventStartDatetime,
		constants.TICKET_STATUS_SETTLED_PENDING_PAYOUT,
		tableID,
		event.ID,
		event.PrevEventID,
	).Find(&tickets)

	if res.Error == gorm.ErrRecordNotFound {
		return nil
	}

	return &tickets
}
