package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/wallissonmarinho/GoTV/internal/adapters/storage"
)

func runMigrateCLI(args []string) int {
	lg := log.New(os.Stderr, "gotv ", log.LstdFlags|log.Lmicroseconds)
	if len(args) < 1 {
		lg.Println("usage: gotv migrate <up|down|version|status|goto> [args]")
		lg.Println("  up              apply all pending migrations")
		lg.Println("  down [N]        roll back N migrations (default 1)")
		lg.Println("  version         print highest applied migration version")
		lg.Println("  status          list migrations and applied state")
		lg.Println("  goto VERSION    migrate up or down to VERSION (0 = schema baseline only)")
		return 2
	}

	dataDir := getenv("GOTV_DATA_DIR", "./data")
	_ = os.MkdirAll(dataDir, 0o755)
	dbPath := filepath.Join(dataDir, "gotv.db")
	sqliteDSN := getenv("GOTV_SQLITE_DSN", "file:"+dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = sqliteDSN
	}

	db, pg, err := storage.OpenDB(dsn)
	if err != nil {
		lg.Println(err)
		return 1
	}
	defer db.Close()

	p, err := storage.NewMigrateProvider(db, pg)
	if err != nil {
		lg.Println(err)
		return 1
	}

	ctx := context.Background()
	cmd := args[0]
	switch cmd {
	case "up":
		_, err = p.Up(ctx)
	case "down":
		n := 1
		if len(args) > 1 {
			n, err = strconv.Atoi(args[1])
			if err != nil || n < 1 {
				lg.Println("down: optional argument must be a positive integer")
				return 2
			}
		}
		err = storage.MigrateDownSteps(ctx, p, n)
	case "version":
		var v int64
		v, err = p.GetDBVersion(ctx)
		if err != nil {
			break
		}
		_, _ = fmt.Fprintf(os.Stdout, "%d\n", v)
	case "status":
		err = storage.PrintMigrateStatus(ctx, p, os.Stdout)
	case "goto":
		if len(args) < 2 {
			lg.Println("goto: missing VERSION")
			return 2
		}
		target, convErr := strconv.ParseInt(args[1], 10, 64)
		if convErr != nil || target < 0 {
			lg.Println("goto: VERSION must be a non-negative integer")
			return 2
		}
		var cur int64
		cur, err = p.GetDBVersion(ctx)
		if err != nil {
			break
		}
		if target > cur {
			_, err = p.UpTo(ctx, target)
		} else if target < cur {
			_, err = p.DownTo(ctx, target)
		}
	default:
		lg.Printf("unknown migrate subcommand %q", cmd)
		return 2
	}

	if err != nil {
		lg.Println(err)
		return 1
	}
	return 0
}
