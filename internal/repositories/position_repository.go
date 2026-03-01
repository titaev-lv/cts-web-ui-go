package repositories

import (
	"ctweb/internal/db"
	"ctweb/internal/models"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type PositionRepository struct{}

func NewPositionRepository() *PositionRepository {
	return &PositionRepository{}
}

func (r *PositionRepository) CountPositionsByUser(userID int) (int, error) {
	query := `SELECT COUNT(*) AS count FROM POS_POSITIONS WHERE USER_ID = ?`
	var count int
	if err := db.DB.QueryRow(query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count positions: %w", err)
	}
	return count, nil
}

func (r *PositionRepository) GetPositions(userID, limit, offset int) ([]*models.PositionSummary, error) {
	query := `WITH RECURSIVE
			ordered AS (
				SELECT
					tt.POSITION_ID,
					tt.ID        AS TRADE_ID,
					tt.PRICE,
					tt.VOLUME,
					tt.FEE,
					tt.FEE_BASE,
					tt.FUNDING_AMOUNT,
					p.MARKET_TYPE,
					tt.OP_TYPE,
					ROW_NUMBER() OVER (PARTITION BY tt.POSITION_ID ORDER BY tt.ID) AS rn
				FROM
					POS_TRANSACTIONS tt
				JOIN
					POS_POSITIONS p
						ON p.ID = tt.POSITION_ID
				WHERE
					p.USER_ID = ?
			),
			trade_calc AS (
				SELECT
					o.POSITION_ID,
					o.TRADE_ID,
					o.PRICE,
					o.VOLUME,
					CASE
						WHEN o.MARKET_TYPE='SPOT' AND o.VOLUME>0
							THEN o.VOLUME-COALESCE(o.FEE_BASE,0)
						ELSE o.VOLUME
					END AS POS,
					CASE
						WHEN o.MARKET_TYPE='FUTURES' AND o.VOLUME<>0
							THEN (o.PRICE*NULLIF(o.VOLUME,0) + COALESCE(o.FEE,0))/NULLIF(o.VOLUME,0)
						WHEN o.MARKET_TYPE='SPOT' AND o.VOLUME<0
							THEN (o.PRICE*NULLIF(o.VOLUME,0) + COALESCE(o.FEE,0))/NULLIF(o.VOLUME,0)
						WHEN o.MARKET_TYPE='SPOT' AND o.VOLUME>0
							THEN o.PRICE*NULLIF(o.VOLUME,0)/NULLIF((o.VOLUME-COALESCE(o.FEE_BASE,0)),0)
						ELSE o.PRICE
					END AS AVG_PRICE,
					CAST(0 AS DECIMAL(32,16)) AS REALIZED_PNL,
					COALESCE(o.FEE,0)          AS FEE_ACCUM,
					COALESCE(o.FEE_BASE,0)     AS FEE_BASE_ACCUM,
					COALESCE(o.FUNDING_AMOUNT,0) AS FUNDING_ACCUM,
					o.MARKET_TYPE,
					o.OP_TYPE,
					o.rn
				FROM
					ordered o
				WHERE
					o.rn = 1
				UNION ALL
				SELECT
					o.POSITION_ID,
					o.TRADE_ID,
					o.PRICE,
					o.VOLUME,
					prev.POS + CASE
									WHEN prev.MARKET_TYPE='SPOT' AND o.VOLUME>0
										THEN o.VOLUME-COALESCE(o.FEE_BASE,0)
									ELSE o.VOLUME
							   END AS POS,
					CASE
						WHEN prev.MARKET_TYPE='FUTURES' AND o.VOLUME<>0
							THEN (prev.POS*IFNULL(prev.AVG_PRICE,0) + o.VOLUME*o.PRICE + o.FEE - prev.REALIZED_PNL) / NULLIF(prev.POS+o.VOLUME,0)
						WHEN prev.MARKET_TYPE='FUTURES' AND o.FUNDING_AMOUNT<>0
							THEN (prev.POS*prev.AVG_PRICE - o.FUNDING_AMOUNT) / NULLIF(prev.POS,0)
						WHEN prev.MARKET_TYPE='SPOT' AND o.VOLUME>0
							THEN (prev.POS*IFNULL(prev.AVG_PRICE,0) + (o.VOLUME-COALESCE(o.FEE_BASE,0))*((o.PRICE*NULLIF(o.VOLUME,0))/NULLIF((o.VOLUME-COALESCE(o.FEE_BASE,0)),0)) - prev.REALIZED_PNL) / NULLIF(prev.POS+o.VOLUME-o.FEE_BASE,0)
						WHEN prev.MARKET_TYPE='SPOT' AND o.VOLUME<0
							THEN (prev.POS*IFNULL(prev.AVG_PRICE,0) + o.VOLUME*o.PRICE + o.FEE - prev.REALIZED_PNL) / NULLIF(prev.POS+o.VOLUME,0)
						ELSE NULL
					END AS AVG_PRICE,
					CASE
						WHEN prev.MARKET_TYPE='FUTURES' OR (prev.MARKET_TYPE='SPOT' AND o.VOLUME<0) THEN
							CASE
								WHEN (prev.POS + o.VOLUME) = 0
									THEN (o.PRICE - prev.AVG_PRICE) * LEAST(ABS(o.VOLUME),ABS(prev.POS)) * SIGN(prev.POS) - o.FEE
								ELSE 0
							END
						WHEN prev.MARKET_TYPE='SPOT' AND o.VOLUME>0 THEN
							CASE
								WHEN (prev.POS + (o.VOLUME-o.FEE_BASE)) = 0
									THEN (o.PRICE - prev.AVG_PRICE) * LEAST(ABS(o.VOLUME),ABS(prev.POS)) * SIGN(prev.POS)
								ELSE 0
							END
						ELSE 0
					END AS REALIZED_PNL,
					prev.FEE_ACCUM       + COALESCE(o.FEE,0)          AS FEE_ACCUM,
					prev.FEE_BASE_ACCUM  + COALESCE(o.FEE_BASE,0)     AS FEE_BASE_ACCUM,
					prev.FUNDING_ACCUM   + COALESCE(o.FUNDING_AMOUNT,0) AS FUNDING_ACCUM,
					prev.MARKET_TYPE,
					prev.OP_TYPE,
					o.rn
				FROM
					trade_calc prev
				JOIN
					ordered o
						ON o.POSITION_ID = prev.POSITION_ID AND o.rn = prev.rn + 1
			),
			last_calc AS (
				SELECT
					tc.*
				FROM
					trade_calc tc
				JOIN (
					SELECT
						POSITION_ID, MAX(rn) AS last_rn
					FROM
						trade_calc
					GROUP BY
						POSITION_ID
				) last_tc
					ON tc.POSITION_ID = last_tc.POSITION_ID AND tc.rn = last_tc.last_rn
			)
			SELECT
				p.ID AS POSITION_ID,
				p.NAME AS CONTRACT_NAME,
				e.NAME AS EXCHANGE_NAME,
				p.MARKET_TYPE,
				CASE
					WHEN p.STATUS = 1
						THEN 'OPEN'
					ELSE 'CLOSE'
				END AS STATUS,
				p.CREATED,
				p.CLOSED,
				CAST(lc.POS AS DOUBLE) AS FINAL_POSITION,
				CAST(CASE WHEN lc.POS=0 THEN NULL ELSE lc.AVG_PRICE END AS DOUBLE) AS FINAL_AVG_PRICE,
				CAST(lc.FEE_BASE_ACCUM AS DOUBLE) AS FEE_BASE_TOTAL,
				CAST(lc.FEE_ACCUM AS DOUBLE) AS FEE_TOTAL,
				CAST(lc.FUNDING_ACCUM AS DOUBLE) AS FUNDING_TOTAL,
				CAST(lc.REALIZED_PNL AS DOUBLE) AS TOTAL_REALIZED_PNL
			FROM
				POS_POSITIONS p
			LEFT JOIN
				last_calc lc ON lc.POSITION_ID = p.ID
			LEFT JOIN
				EXCHANGE e   ON e.ID = p.EXID
			WHERE
				p.USER_ID = ?
			ORDER BY
				(p.STATUS='OPEN') ASC,
				p.CREATED DESC
			LIMIT ? OFFSET ?`

	rows, err := db.DB.Query(query, userID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}
	defer rows.Close()

	result := make([]*models.PositionSummary, 0)
	for rows.Next() {
		var item models.PositionSummary
		var created sql.NullTime
		var closed sql.NullTime
		var finalPos sql.NullFloat64
		var finalAvg sql.NullFloat64
		var feeBase sql.NullFloat64
		var fee sql.NullFloat64
		var funding sql.NullFloat64
		var realized sql.NullFloat64

		err = rows.Scan(
			&item.PositionID,
			&item.ContractName,
			&item.ExchangeName,
			&item.MarketType,
			&item.Status,
			&created,
			&closed,
			&finalPos,
			&finalAvg,
			&feeBase,
			&fee,
			&funding,
			&realized,
		)
		if err != nil {
			return nil, fmt.Errorf("scan positions row: %w", err)
		}

		if created.Valid {
			item.Created = &created.Time
		}
		if closed.Valid {
			item.Closed = &closed.Time
		}
		if finalPos.Valid {
			item.FinalPosition = &finalPos.Float64
		}
		if finalAvg.Valid {
			item.FinalAvgPrice = &finalAvg.Float64
		}
		if feeBase.Valid {
			item.FeeBaseTotal = &feeBase.Float64
		}
		if fee.Valid {
			item.FeeTotal = &fee.Float64
		}
		if funding.Valid {
			item.FundingTotal = &funding.Float64
		}
		if realized.Valid {
			item.TotalRealizedPnL = &realized.Float64
		}

		result = append(result, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate positions rows: %w", err)
	}

	return result, nil
}

func (r *PositionRepository) CreatePosition(name string, exchangeID int, createdUTC time.Time, market string, userID int) error {
	query := `INSERT INTO POS_POSITIONS (NAME, EXID, CREATED, MARKET_TYPE, USER_ID) VALUES(?,?,?,?,?)`
	res, err := db.DB.Exec(query, name, exchangeID, createdUTC.Format("2006-01-02 15:04:05"), market, userID)
	if err != nil {
		return fmt.Errorf("create position: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("create position rows affected: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("create position failed")
	}

	return nil
}

func (r *PositionRepository) EditPosition(positionID, userID int, name string, exchangeID int, createdUTC time.Time) (bool, error) {
	query := `UPDATE POS_POSITIONS SET NAME = ?, EXID = ?, CREATED = ? WHERE USER_ID = ? AND ID = ?`
	res, err := db.DB.Exec(query, name, exchangeID, createdUTC.Format("2006-01-02 15:04:05"), userID, positionID)
	if err != nil {
		return false, fmt.Errorf("edit position: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("edit position rows affected: %w", err)
	}

	return affected > 0, nil
}

func (r *PositionRepository) GetPositionByID(userID, positionID int) (*models.PositionDetail, error) {
	query := `WITH RECURSIVE
			ordered AS (
				SELECT
					tt.POSITION_ID,
					tt.ID        AS TRADE_ID,
					tt.PRICE,
					tt.VOLUME,
					tt.FEE,
					tt.FEE_BASE,
					tt.FUNDING_AMOUNT,
					p.MARKET_TYPE,
					tt.OP_TYPE,
					ROW_NUMBER() OVER (PARTITION BY tt.POSITION_ID ORDER BY tt.ID) AS rn
				FROM
					POS_TRANSACTIONS tt
				JOIN
					POS_POSITIONS p
						ON p.ID = tt.POSITION_ID
				WHERE
					p.USER_ID = ?
					AND p.ID= ?
			),
			trade_calc AS (
				SELECT
					o.POSITION_ID,
					o.TRADE_ID,
					o.PRICE,
					o.VOLUME,
					CASE
						WHEN o.MARKET_TYPE='SPOT' AND o.VOLUME>0
							THEN o.VOLUME-COALESCE(o.FEE_BASE,0)
						ELSE o.VOLUME
					END AS POS,
					CASE
						WHEN o.MARKET_TYPE='FUTURES' AND o.VOLUME<>0
							THEN (o.PRICE*NULLIF(o.VOLUME,0) + COALESCE(o.FEE,0))/NULLIF(o.VOLUME,0)
						WHEN o.MARKET_TYPE='SPOT' AND o.VOLUME<0
							THEN (o.PRICE*NULLIF(o.VOLUME,0) + COALESCE(o.FEE,0))/NULLIF(o.VOLUME,0)
						WHEN o.MARKET_TYPE='SPOT' AND o.VOLUME>0
							THEN o.PRICE*NULLIF(o.VOLUME,0)/NULLIF((o.VOLUME-COALESCE(o.FEE_BASE,0)),0)
						ELSE o.PRICE
					END AS AVG_PRICE,
					CAST(0 AS DECIMAL(32,16)) AS REALIZED_PNL,
					COALESCE(o.FEE,0)          AS FEE_ACCUM,
					COALESCE(o.FEE_BASE,0)     AS FEE_BASE_ACCUM,
					COALESCE(o.FUNDING_AMOUNT,0) AS FUNDING_ACCUM,
					o.MARKET_TYPE,
					o.OP_TYPE,
					o.rn
				FROM
					ordered o
				WHERE
					o.rn = 1
				UNION ALL
				SELECT
					o.POSITION_ID,
					o.TRADE_ID,
					o.PRICE,
					o.VOLUME,
					prev.POS + CASE
									WHEN prev.MARKET_TYPE='SPOT' AND o.VOLUME>0
										THEN o.VOLUME-COALESCE(o.FEE_BASE,0)
									ELSE o.VOLUME
							   END AS POS,
					CASE
						WHEN prev.MARKET_TYPE='FUTURES' AND o.VOLUME<>0
							THEN (prev.POS*IFNULL(prev.AVG_PRICE,0) + o.VOLUME*o.PRICE + o.FEE - prev.REALIZED_PNL) / NULLIF(prev.POS+o.VOLUME,0)
						WHEN prev.MARKET_TYPE='FUTURES' AND o.FUNDING_AMOUNT<>0
							THEN (prev.POS*prev.AVG_PRICE - o.FUNDING_AMOUNT) / NULLIF(prev.POS,0)
						WHEN prev.MARKET_TYPE='SPOT' AND o.VOLUME>0
							THEN (prev.POS*IFNULL(prev.AVG_PRICE,0) + (o.VOLUME-COALESCE(o.FEE_BASE,0))*((o.PRICE*NULLIF(o.VOLUME,0))/NULLIF((o.VOLUME-COALESCE(o.FEE_BASE,0)),0)) - prev.REALIZED_PNL) / NULLIF(prev.POS+o.VOLUME-o.FEE_BASE,0)
						WHEN prev.MARKET_TYPE='SPOT' AND o.VOLUME<0
							THEN (prev.POS*IFNULL(prev.AVG_PRICE,0) + o.VOLUME*o.PRICE + o.FEE - prev.REALIZED_PNL) / NULLIF(prev.POS+o.VOLUME,0)
						ELSE NULL
					END AS AVG_PRICE,
					CASE
						WHEN prev.MARKET_TYPE='FUTURES' OR (prev.MARKET_TYPE='SPOT' AND o.VOLUME<0) THEN
							CASE
								WHEN (prev.POS + o.VOLUME) = 0
									THEN (o.PRICE - prev.AVG_PRICE) * LEAST(ABS(o.VOLUME),ABS(prev.POS)) * SIGN(prev.POS) - o.FEE
								ELSE 0
							END
						WHEN prev.MARKET_TYPE='SPOT' AND o.VOLUME>0 THEN
							CASE
								WHEN (prev.POS + (o.VOLUME-o.FEE_BASE)) = 0
									THEN (o.PRICE - prev.AVG_PRICE) * LEAST(ABS(o.VOLUME),ABS(prev.POS)) * SIGN(prev.POS)
								ELSE 0
							END
						ELSE 0
					END AS REALIZED_PNL,
					prev.FEE_ACCUM       + COALESCE(o.FEE,0)          AS FEE_ACCUM,
					prev.FEE_BASE_ACCUM  + COALESCE(o.FEE_BASE,0)     AS FEE_BASE_ACCUM,
					prev.FUNDING_ACCUM   + COALESCE(o.FUNDING_AMOUNT,0) AS FUNDING_ACCUM,
					prev.MARKET_TYPE,
					prev.OP_TYPE,
					o.rn
				FROM
					trade_calc prev
				JOIN
					ordered o
						ON o.POSITION_ID = prev.POSITION_ID AND o.rn = prev.rn + 1
			),
			last_calc AS (
				SELECT
					tc.*
				FROM
					trade_calc tc
				JOIN (
					SELECT
						POSITION_ID, MAX(rn) AS last_rn
					FROM
						trade_calc
					GROUP BY
						POSITION_ID
				) last_tc
					ON tc.POSITION_ID = last_tc.POSITION_ID AND tc.rn = last_tc.last_rn
			)
			SELECT
				p.ID AS POSITION_ID,
				p.NAME AS CONTRACT_NAME,
				e.NAME AS EXCHANGE_NAME,
				p.MARKET_TYPE,
				CASE
					WHEN p.STATUS = 1
						THEN 'OPEN'
					ELSE 'CLOSE'
				END AS STATUS,
				p.CREATED,
				p.CLOSED,
				CAST(lc.POS AS DOUBLE) AS FINAL_POSITION,
				CAST(CASE WHEN lc.POS=0 THEN NULL ELSE lc.AVG_PRICE END AS DOUBLE) AS FINAL_AVG_PRICE,
				CAST(lc.FEE_BASE_ACCUM AS DOUBLE) AS FEE_BASE_TOTAL,
				CAST(lc.FEE_ACCUM AS DOUBLE) AS FEE_TOTAL,
				CAST(lc.FUNDING_ACCUM AS DOUBLE) AS FUNDING_TOTAL,
				CAST(lc.REALIZED_PNL AS DOUBLE) AS TOTAL_REALIZED_PNL,
				rn
			FROM
				POS_POSITIONS p
			LEFT JOIN
				last_calc lc ON lc.POSITION_ID = p.ID
			LEFT JOIN
				EXCHANGE e   ON e.ID = p.EXID
			WHERE
				p.USER_ID = ?
				AND p.ID= ?
			ORDER BY
				(p.STATUS='OPEN') DESC,
				p.CREATED DESC`

	var item models.PositionDetail
	var created sql.NullTime
	var closed sql.NullTime
	var finalPos sql.NullFloat64
	var finalAvg sql.NullFloat64
	var feeBase sql.NullFloat64
	var fee sql.NullFloat64
	var funding sql.NullFloat64
	var realized sql.NullFloat64
	var transCount sql.NullInt64

	err := db.DB.QueryRow(query, userID, positionID, userID, positionID).Scan(
		&item.PositionID,
		&item.ContractName,
		&item.ExchangeName,
		&item.MarketType,
		&item.Status,
		&created,
		&closed,
		&finalPos,
		&finalAvg,
		&feeBase,
		&fee,
		&funding,
		&realized,
		&transCount,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get position by id: %w", err)
	}

	if created.Valid {
		item.Created = &created.Time
	}
	if closed.Valid {
		item.Closed = &closed.Time
	}
	if finalPos.Valid {
		item.FinalPosition = &finalPos.Float64
	}
	if finalAvg.Valid {
		item.FinalAvgPrice = &finalAvg.Float64
	}
	if feeBase.Valid {
		item.FeeBaseTotal = &feeBase.Float64
	}
	if fee.Valid {
		item.FeeTotal = &fee.Float64
	}
	if funding.Valid {
		item.FundingTotal = &funding.Float64
	}
	if realized.Valid {
		item.TotalRealizedPnL = &realized.Float64
	}
	if transCount.Valid {
		item.TransCount = int(transCount.Int64)
	}

	return &item, nil
}

func (r *PositionRepository) CountTransactionsByPosition(positionID, userID int) (int, error) {
	query := `SELECT
				COUNT(*) AS CNT
			FROM
				POS_TRANSACTIONS t
			JOIN
				POS_POSITIONS pp ON pp.ID = t.POSITION_ID
			JOIN
				USER u ON u.ID = pp.USER_ID
			WHERE
				t.POSITION_ID = ?
				AND u.ID = ?`

	var count int
	if err := db.DB.QueryRow(query, positionID, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count transactions: %w", err)
	}

	return count, nil
}

func (r *PositionRepository) GetTransactionsByPosition(positionID, userID, limit, offset int) ([]*models.PositionTransaction, error) {
	query := `SELECT
				t.ID,
				t.OP_TYPE AS TYPE,
				CAST(t.PRICE AS DOUBLE) AS PRICE,
				CAST(t.VOLUME AS DOUBLE) AS VOLUME,
				CAST(t.FEE_BASE AS DOUBLE) AS FEE_BASE,
				CAST(t.FEE AS DOUBLE) AS FEE,
				CAST(t.FUNDING_AMOUNT AS DOUBLE) AS FUNDING,
				t.TRANS_DATE
			FROM
				POS_TRANSACTIONS t
			JOIN
				POS_POSITIONS pp ON pp.ID = t.POSITION_ID
			JOIN
				USER u ON u.ID = pp.USER_ID
			WHERE
				t.POSITION_ID = ?
				AND u.ID = ?
			ORDER BY
				t.TRANS_DATE DESC,
				t.ID DESC
			LIMIT ? OFFSET ?`

	rows, err := db.DB.Query(query, positionID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}
	defer rows.Close()

	result := make([]*models.PositionTransaction, 0)
	for rows.Next() {
		var item models.PositionTransaction
		var transDate sql.NullTime
		err = rows.Scan(
			&item.ID,
			&item.Type,
			&item.Price,
			&item.Volume,
			&item.FeeBase,
			&item.Fee,
			&item.Funding,
			&transDate,
		)
		if err != nil {
			return nil, fmt.Errorf("scan transactions row: %w", err)
		}
		if transDate.Valid {
			item.TransDate = &transDate.Time
		}
		result = append(result, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions rows: %w", err)
	}

	return result, nil
}

func (r *PositionRepository) GetPositionMarketType(positionID, userID int) (string, error) {
	query := `SELECT MARKET_TYPE FROM POS_POSITIONS WHERE ID = ? AND USER_ID = ?`
	var marketType string
	if err := db.DB.QueryRow(query, positionID, userID).Scan(&marketType); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("get position market type: %w", err)
	}
	return marketType, nil
}

func (r *PositionRepository) InsertFundingTransaction(positionID int, funding float64, transDateUTC time.Time) error {
	query := `INSERT INTO POS_TRANSACTIONS (POSITION_ID, FUNDING_AMOUNT, TRANS_DATE, OP_TYPE) VALUES(?,?,?,?)`
	_, err := db.DB.Exec(query, positionID, funding, transDateUTC.Format("2006-01-02 15:04:05"), "FUNDING")
	if err != nil {
		return fmt.Errorf("insert funding transaction: %w", err)
	}
	return nil
}

func (r *PositionRepository) InsertFundingTransactionRaw(positionID int, funding string, transDateUTC time.Time) error {
	query := `INSERT INTO POS_TRANSACTIONS (POSITION_ID, FUNDING_AMOUNT, TRANS_DATE, OP_TYPE) VALUES(?,?,?,?)`
	_, err := db.DB.Exec(query, positionID, funding, transDateUTC.Format("2006-01-02 15:04:05.000"), "FUNDING")
	if err != nil {
		return fmt.Errorf("insert funding transaction raw: %w", err)
	}
	return nil
}

func (r *PositionRepository) InsertFundingTransactionImport(positionID int, funding string, transDateUTC time.Time, sourceOrderID, sourceTradeID *string) (bool, error) {
	query := `INSERT IGNORE INTO POS_TRANSACTIONS (POSITION_ID, FUNDING_AMOUNT, TRANS_DATE, OP_TYPE, SOURCE_ORDER_ID, SOURCE_TRADE_ID) VALUES(?,?,?,?,?,?)`
	res, err := db.DB.Exec(query, positionID, funding, transDateUTC.Format("2006-01-02 15:04:05.000"), "FUNDING", sourceOrderID, sourceTradeID)
	if err != nil {
		return false, fmt.Errorf("insert funding transaction import: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("insert funding transaction import rows affected: %w", err)
	}

	return affected > 0, nil
}

func (r *PositionRepository) InsertTradeTransaction(positionID int, price, volume, fee, feeBase float64, transDateUTC time.Time) error {
	if feeBase != 0 {
		query := `INSERT INTO POS_TRANSACTIONS (POSITION_ID, PRICE, VOLUME, FEE_BASE, TRANS_DATE, OP_TYPE) VALUES(?,?,?,?,?,?)`
		_, err := db.DB.Exec(query, positionID, price, volume, feeBase, transDateUTC.Format("2006-01-02 15:04:05"), "TRADE")
		if err != nil {
			return fmt.Errorf("insert spot buy transaction: %w", err)
		}
		return nil
	}

	query := `INSERT INTO POS_TRANSACTIONS (POSITION_ID, PRICE, VOLUME, FEE, TRANS_DATE, OP_TYPE) VALUES(?,?,?,?,?,?)`
	_, err := db.DB.Exec(query, positionID, price, volume, fee, transDateUTC.Format("2006-01-02 15:04:05"), "TRADE")
	if err != nil {
		return fmt.Errorf("insert trade transaction: %w", err)
	}
	return nil
}

func (r *PositionRepository) InsertTradeTransactionRaw(positionID int, price, volume, fee, feeBase string, transDateUTC time.Time) error {
	if feeBase != "0" && feeBase != "0.0" && feeBase != "0.00" && feeBase != "0.00000000000000000000" {
		query := `INSERT INTO POS_TRANSACTIONS (POSITION_ID, PRICE, VOLUME, FEE_BASE, TRANS_DATE, OP_TYPE) VALUES(?,?,?,?,?,?)`
		_, err := db.DB.Exec(query, positionID, price, volume, feeBase, transDateUTC.Format("2006-01-02 15:04:05.000"), "TRADE")
		if err != nil {
			return fmt.Errorf("insert spot buy transaction raw: %w", err)
		}
		return nil
	}

	query := `INSERT INTO POS_TRANSACTIONS (POSITION_ID, PRICE, VOLUME, FEE, TRANS_DATE, OP_TYPE) VALUES(?,?,?,?,?,?)`
	_, err := db.DB.Exec(query, positionID, price, volume, fee, transDateUTC.Format("2006-01-02 15:04:05.000"), "TRADE")
	if err != nil {
		return fmt.Errorf("insert trade transaction raw: %w", err)
	}
	return nil
}

func (r *PositionRepository) InsertTradeTransactionImport(positionID int, price, volume, fee, feeBase string, transDateUTC time.Time, sourceOrderID, sourceTradeID *string) (bool, error) {
	if feeBase != "0" && feeBase != "0.0" && feeBase != "0.00" && feeBase != "0.00000000000000000000" {
		query := `INSERT IGNORE INTO POS_TRANSACTIONS (POSITION_ID, PRICE, VOLUME, FEE_BASE, TRANS_DATE, OP_TYPE, SOURCE_ORDER_ID, SOURCE_TRADE_ID) VALUES(?,?,?,?,?,?,?,?)`
		res, err := db.DB.Exec(query, positionID, price, volume, feeBase, transDateUTC.Format("2006-01-02 15:04:05.000"), "TRADE", sourceOrderID, sourceTradeID)
		if err != nil {
			return false, fmt.Errorf("insert spot buy transaction import: %w", err)
		}

		affected, err := res.RowsAffected()
		if err != nil {
			return false, fmt.Errorf("insert spot buy transaction import rows affected: %w", err)
		}

		return affected > 0, nil
	}

	query := `INSERT IGNORE INTO POS_TRANSACTIONS (POSITION_ID, PRICE, VOLUME, FEE, TRANS_DATE, OP_TYPE, SOURCE_ORDER_ID, SOURCE_TRADE_ID) VALUES(?,?,?,?,?,?,?,?)`
	res, err := db.DB.Exec(query, positionID, price, volume, fee, transDateUTC.Format("2006-01-02 15:04:05.000"), "TRADE", sourceOrderID, sourceTradeID)
	if err != nil {
		return false, fmt.Errorf("insert trade transaction import: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("insert trade transaction import rows affected: %w", err)
	}

	return affected > 0, nil
}

func (r *PositionRepository) GetTransactionByID(userID, positionID, transactionID int) (*models.PositionTransaction, error) {
	query := `SELECT
				t.ID,
				t.OP_TYPE AS TYPE,
				CAST(t.PRICE AS DOUBLE) AS PRICE,
				CAST(t.VOLUME AS DOUBLE) AS VOLUME,
				CAST(t.FEE_BASE AS DOUBLE) AS FEE_BASE,
				CAST(t.FEE AS DOUBLE) AS FEE,
				CAST(t.FUNDING_AMOUNT AS DOUBLE) AS FUNDING,
				t.TRANS_DATE
			FROM
				POS_TRANSACTIONS t
			JOIN
				POS_POSITIONS pp ON pp.ID = t.POSITION_ID
			WHERE
				pp.USER_ID = ?
				AND pp.ID = ?
				AND t.ID = ?
			LIMIT 1`

	var item models.PositionTransaction
	var transDate sql.NullTime
	err := db.DB.QueryRow(query, userID, positionID, transactionID).Scan(
		&item.ID,
		&item.Type,
		&item.Price,
		&item.Volume,
		&item.FeeBase,
		&item.Fee,
		&item.Funding,
		&transDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}

	if transDate.Valid {
		item.TransDate = &transDate.Time
	}

	return &item, nil
}

func (r *PositionRepository) UpdateTransactionByID(userID, positionID, transactionID int, price, volume, fee, feeBase, funding float64, transDateUTC time.Time, opType string) (bool, error) {
	query := `UPDATE POS_TRANSACTIONS t
			JOIN POS_POSITIONS p ON p.ID = t.POSITION_ID
			SET
				t.PRICE = ?,
				t.VOLUME = ?,
				t.FEE = ?,
				t.FEE_BASE = ?,
				t.FUNDING_AMOUNT = ?,
				t.TRANS_DATE = ?,
				t.OP_TYPE = ?
			WHERE
				p.USER_ID = ?
				AND p.ID = ?
				AND t.ID = ?`

	res, err := db.DB.Exec(
		query,
		price,
		volume,
		fee,
		feeBase,
		funding,
		transDateUTC.Format("2006-01-02 15:04:05"),
		opType,
		userID,
		positionID,
		transactionID,
	)
	if err != nil {
		return false, fmt.Errorf("update transaction: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update transaction rows affected: %w", err)
	}

	if affected == 0 {
		existsQuery := `SELECT 1
						FROM POS_TRANSACTIONS t
						JOIN POS_POSITIONS p ON p.ID = t.POSITION_ID
						WHERE p.USER_ID = ?
						  AND p.ID = ?
						  AND t.ID = ?
						LIMIT 1`

		var marker int
		err := db.DB.QueryRow(existsQuery, userID, positionID, transactionID).Scan(&marker)
		if err == sql.ErrNoRows {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("check transaction existence: %w", err)
		}

		return true, nil
	}

	return affected > 0, nil
}

func (r *PositionRepository) GetPositionStatusAndAmount(positionID, userID int) (string, float64, error) {
	query := `WITH RECURSIVE
			ordered AS (
				SELECT
					tt.POSITION_ID,
					tt.ID        AS TRADE_ID,
					tt.VOLUME,
					tt.FEE,
					tt.FEE_BASE,
					p.MARKET_TYPE,
					ROW_NUMBER() OVER (PARTITION BY tt.POSITION_ID ORDER BY tt.ID) AS rn
				FROM
					POS_TRANSACTIONS tt
				JOIN
					POS_POSITIONS p
						ON p.ID = tt.POSITION_ID
				WHERE
					p.USER_ID = ?
					AND p.ID = ?
			),
			trade_calc AS (
				SELECT
					o.POSITION_ID,
					o.TRADE_ID,
					o.VOLUME,
					CASE
						WHEN o.MARKET_TYPE='SPOT' AND o.VOLUME>0
							THEN o.VOLUME-COALESCE(o.FEE_BASE,0)
						ELSE o.VOLUME
					END AS POS,
					o.FEE_BASE,
					o.FEE,
					o.MARKET_TYPE,
					o.rn
				FROM
					ordered o
				WHERE
					o.rn = 1
				UNION ALL
				SELECT
					o.POSITION_ID,
					o.TRADE_ID,
					o.VOLUME,
					prev.POS + CASE
									WHEN prev.MARKET_TYPE='SPOT' AND o.VOLUME>0
										THEN o.VOLUME-COALESCE(o.FEE_BASE,0)
									ELSE o.VOLUME
							   END AS POS,
					o.FEE_BASE,
					o.FEE,
					prev.MARKET_TYPE,
					o.rn
				FROM
					trade_calc prev
				JOIN
					ordered o
						ON o.POSITION_ID = prev.POSITION_ID AND o.rn = prev.rn + 1
			),
			last_calc AS (
				SELECT
					tc.*
				FROM
					trade_calc tc
				JOIN (
					SELECT
						POSITION_ID, MAX(rn) AS last_rn
					FROM
						trade_calc
					GROUP BY
						POSITION_ID
				) last_tc
					ON tc.POSITION_ID = last_tc.POSITION_ID AND tc.rn = last_tc.last_rn
			)
			SELECT
				CASE
					WHEN p.STATUS = 1
						THEN 'OPEN'
					ELSE 'CLOSE'
				END AS STATUS,
				CAST(lc.POS AS DOUBLE) AS POSITION
			FROM
				POS_POSITIONS p
			LEFT JOIN
				last_calc lc ON lc.POSITION_ID = p.ID
			WHERE
				p.USER_ID = ?
				AND p.ID = ?`

	var status sql.NullString
	var amount sql.NullFloat64
	err := db.DB.QueryRow(query, userID, positionID, userID, positionID).Scan(&status, &amount)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, nil
		}
		return "", 0, fmt.Errorf("get position status and amount: %w", err)
	}

	statusValue := ""
	if status.Valid {
		statusValue = status.String
	}
	amountValue := 0.0
	if amount.Valid {
		amountValue = amount.Float64
	}

	return statusValue, amountValue, nil
}

func (r *PositionRepository) ClosePosition(positionID, userID int) (bool, error) {
	query := `UPDATE POS_POSITIONS SET STATUS = 0, CLOSED = NOW() WHERE USER_ID = ? AND ID = ?`
	res, err := db.DB.Exec(query, userID, positionID)
	if err != nil {
		return false, fmt.Errorf("close position: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("close position rows affected: %w", err)
	}

	return affected > 0, nil
}

func (r *PositionRepository) DeletePosition(positionID, userID int) (bool, error) {
	query := `DELETE FROM POS_POSITIONS WHERE USER_ID = ? AND ID = ?`
	res, err := db.DB.Exec(query, userID, positionID)
	if err != nil {
		return false, fmt.Errorf("delete position: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete position rows affected: %w", err)
	}

	return affected > 0, nil
}

func (r *PositionRepository) DeleteTransactionsByIDs(userID, positionID int, transactionIDs []int) (int64, error) {
	if len(transactionIDs) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(transactionIDs))
	args := make([]interface{}, 0, 2+len(transactionIDs))
	args = append(args, userID, positionID)
	for i, id := range transactionIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `DELETE t
			  FROM POS_TRANSACTIONS t
			  JOIN POS_POSITIONS p ON p.ID = t.POSITION_ID
			 WHERE p.USER_ID = ?
			   AND p.ID = ?
			   AND t.ID IN (` + strings.Join(placeholders, ",") + `)`

	res, err := db.DB.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("delete transactions: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete transactions rows affected: %w", err)
	}

	return affected, nil
}
