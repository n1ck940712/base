package validate

import (
	"testing"
)

type Agent struct {
	name     string
	Alias    string
	ptrRole  *string
	subAgent *Agent
}

func TestValidate(t *testing.T) {
	executeTest(func(err error) {
		t.Fatal(err)
	})
}

func BenchmarkValidate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		executeTest(func(err error) {
			b.Fatal(err)
		})
	}
}

func executeTest(errCallback func(error)) {
	role := "TEST ROLE"
	subAgent := &Agent{
		name:    "Sub Kamote2",
		Alias:   "Sub Sweet potato",
		ptrRole: &role,
	}

	yes := true
	typeObj := map[string]any{
		"type":        "bet",
		"market_type": 18,
		"amount":      5.0,
		"testnil":     nil,
		"bool":        &yes,
		"ticket": &map[string]any{
			"id":     "TicketID123",
			"source": "test.go",
			"urls":   []string{"https://www.foufos.com", "https://www.google.com", "https://www.pornhub.com"},
			"is_nil": nil,
			"tags":   []string{"bet", "tickets", "test"},
			"struct": Agent{
				name:     "Kamote",
				Alias:    "Sweet potato",
				ptrRole:  nil,
				subAgent: subAgent,
			},
			"bets": []map[string]any{
				{
					"id":     123,
					"amount": 5.0,
					"odds":   1.56,
					"ref_no": "123",
					"results": []map[string]any{
						{
							"levels": 1,
							"items":  []string{"2", "3"},
						},
						{
							"levels": 4,
							"items":  []string{"5", "7"},
						},
					},
				},
				{
					"id":     456,
					"amount": 5.4,
					"odds":   2.01,
					"ref_no": "456",
					"results": []map[string]any{
						{
							"levels": 3,
							"items":  []string{"1", "3"},
						},
						{
							"levels": 2,
							"items":  []string{"4", "6"},
						},
					},
				},
			},
			"result": map[string]any{
				"event_id": 1234455,
				"levels": map[string]any{
					"level":  1,
					"result": "bomb1",
				},
			},
		},
	}

	if err := Compose(typeObj).
		Type("", Map).
		Value("type", "bet", "selection", "ticket").
		Value("market_type", 18).
		Type("ticket", Map).
		Type("amount", Float).
		Type("testnil", Float, Nil).
		Value("bool", false, true).
		Type("ticket.tags", Array).
		Type("ticket.tags[]", String).
		Value("ticket.tags", []string{"bet", "tickets", "test"}).
		Value("ticket.tags[]", "bet", "tickets", "test").
		Type("ticket.bets", Array).
		Value("ticket.source", "test.go").
		Regex("ticket.source", "^te(\\S)*.go").
		Regex("ticket.urls[]", "^http(\\S)*.com").
		Value("ticket.is_nil", nil).
		Value("ticket.struct", Agent{
			name:     "Kamote",
			Alias:    "Sweet potato",
			ptrRole:  nil,
			subAgent: subAgent,
		}).
		Value("ticket.struct.name", "Kamote").
		Value("ticket.struct.Alias", "Sweet potato").
		Value("ticket.struct.subAgent.name", "Sub Kamote", "Sub Kamote1", "Sub Kamote2").
		Value("ticket.struct.subAgent", *subAgent).
		Type("ticket.struct.subAgent.ptrRole", String).
		Value("ticket.struct.subAgent.ptrRole", nil, role).
		Type("ticket.result.levels.result", String).
		Value("ticket.result.levels.result", "bomb1", "bomb2", "bomb3").
		Type("ticket.bets.id", String, Int).
		Type("ticket.bets.amount", Float).
		Type("ticket.bets.ref_no", String).
		Type("ticket.bets.results", Array).
		Type("ticket.bets.results[]", Map).
		Type("ticket.bets.results.items", Array).
		Type("ticket.bets.results.items[]", String).
		Type("ticket.bets.results.levels", Int).
		Value("ticket.bets.results.levels", 1, 2, 3, 4).
		Check(); err != nil {
		errCallback(err)
	}
	if err := Compose([]string{"1", "2", "3", "4"}).
		Type("", Array).
		Value("[]", "1", "2", "3", "4").
		Check(); err != nil {
		errCallback(err)
	}
}
