package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/cursorworkshop/cursor-gastown/internal/events"
	"github.com/cursorworkshop/cursor-gastown/internal/style"
	"github.com/cursorworkshop/cursor-gastown/internal/workspace"
)

var (
	seanceRole   string
	seanceRig    string
	seanceRecent int
	seanceJSON   bool
)

var seanceCmd = &cobra.Command{
	Use:     "seance",
	GroupID: GroupDiag,
	Short:   "Talk to your predecessor sessions",
Long: `Seance lets you list predecessor sessions.

"Where did you put the stuff you left for me?" - The #1 handoff question.

DISCOVERY:
  gt seance                     # List recent sessions from events
  gt seance --role crew         # Filter by role type
  gt seance --rig gastown       # Filter by rig
  gt seance --recent 10         # Last N sessions

Sessions are discovered from:
  1. Events emitted by SessionStart hooks (~/gt/.events.jsonl)
  2. The [GAS TOWN] beacon makes sessions searchable in /resume`,
	RunE: runSeance,
}

func init() {
	seanceCmd.Flags().StringVar(&seanceRole, "role", "", "Filter by role (crew, polecat, witness, etc.)")
	seanceCmd.Flags().StringVar(&seanceRig, "rig", "", "Filter by rig name")
	seanceCmd.Flags().IntVarP(&seanceRecent, "recent", "n", 20, "Number of recent sessions to show")
	seanceCmd.Flags().BoolVar(&seanceJSON, "json", false, "Output as JSON")

	rootCmd.AddCommand(seanceCmd)
}

// sessionEvent represents a session_start event from our event stream.
type sessionEvent struct {
	Timestamp string                 `json:"ts"`
	Type      string                 `json:"type"`
	Actor     string                 `json:"actor"`
	Payload   map[string]interface{} `json:"payload"`
}

func runSeance(cmd *cobra.Command, args []string) error {
	// Otherwise, list discoverable sessions
	return runSeanceList()
}

func runSeanceList() error {
	townRoot, err := workspace.FindFromCwd()
	if err != nil || townRoot == "" {
		return fmt.Errorf("not in a Gas Town workspace")
	}

	// Read session events from our event stream
	sessions, err := discoverSessions(townRoot)
	if err != nil {
		return fmt.Errorf("discovering sessions: %w", err)
	}

	// Apply filters
	var filtered []sessionEvent
	for _, s := range sessions {
		if seanceRole != "" {
			actor := strings.ToLower(s.Actor)
			if !strings.Contains(actor, strings.ToLower(seanceRole)) {
				continue
			}
		}
		if seanceRig != "" {
			actor := strings.ToLower(s.Actor)
			if !strings.Contains(actor, strings.ToLower(seanceRig)) {
				continue
			}
		}
		filtered = append(filtered, s)
	}

	// Apply limit
	if seanceRecent > 0 && len(filtered) > seanceRecent {
		filtered = filtered[:seanceRecent]
	}

	if seanceJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(filtered)
	}

	if len(filtered) == 0 {
		fmt.Println("No session events found.")
		fmt.Println(style.Dim.Render("Sessions are discovered from ~/gt/.events.jsonl"))
		fmt.Println(style.Dim.Render("Ensure SessionStart hooks emit session_start events"))
		return nil
	}

	// Print header
	fmt.Printf("%s\n\n", style.Bold.Render("Discoverable Sessions"))

	// Column widths
	idWidth := 12
	roleWidth := 26
	timeWidth := 16
	topicWidth := 28

	fmt.Printf("%-*s  %-*s  %-*s  %-*s\n",
		idWidth, "SESSION_ID",
		roleWidth, "ROLE",
		timeWidth, "STARTED",
		topicWidth, "TOPIC")
	fmt.Printf("%s\n", strings.Repeat("─", idWidth+roleWidth+timeWidth+topicWidth+6))

	for _, s := range filtered {
		sessionID := getPayloadString(s.Payload, "session_id")
		if len(sessionID) > idWidth {
			sessionID = sessionID[:idWidth-1] + "…"
		}

		role := s.Actor
		if len(role) > roleWidth {
			role = role[:roleWidth-1] + "…"
		}

		timeStr := formatEventTime(s.Timestamp)

		topic := getPayloadString(s.Payload, "topic")
		if topic == "" {
			topic = "-"
		}
		if len(topic) > topicWidth {
			topic = topic[:topicWidth-1] + "…"
		}

		fmt.Printf("%-*s  %-*s  %-*s  %-*s\n",
			idWidth, sessionID,
			roleWidth, role,
			timeWidth, timeStr,
			topicWidth, topic)
	}

	return nil
}

// discoverSessions reads session_start events from our event stream.
func discoverSessions(townRoot string) ([]sessionEvent, error) {
	eventsPath := filepath.Join(townRoot, events.EventsFile)

	file, err := os.Open(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var sessions []sessionEvent
	scanner := bufio.NewScanner(file)

	// Increase buffer for large lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var event sessionEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}

		if event.Type == events.TypeSessionStart {
			sessions = append(sessions, event)
		}
	}

	// Sort by timestamp descending (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Timestamp > sessions[j].Timestamp
	})

	return sessions, scanner.Err()
}

func getPayloadString(payload map[string]interface{}, key string) string {
	if v, ok := payload[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func formatEventTime(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Local().Format("2006-01-02 15:04")
}
