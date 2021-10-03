package clickhouse
//
import (
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	apiPb "github.com/squzy/squzy_generated/generated/proto/v1"
)

type TransactionInfo struct {
	gorm.Model
	TransactionId     string
	ApplicationId     string
	ParentId          string
	MetaHost          string
	MetaPath          string
	MetaMethod        string
	Name              string
	StartTime         int64
	EndTime           int64
	TransactionStatus int32
	TransactionType   int32
	Error             string
}

type GroupResult struct {
	Name         string
	Count        int64
	SuccessCount int64
	Latency      string
	MinTime      string
	MaxTime      string
	LowTime      string
}

const (
	transNameStr            = "name"
	transMetaHostStr        = "metaHost"
	transMetaMethodStr      = "metaMethod"
	transMetaPathStr        = "metaPath"
	transTransactionTypeStr = "transactionType"
)

var (
	applicationIdFilterString        = fmt.Sprintf(`"%s"."applicationId" = ?`, dbTransactionInfoCollection)
	applicationStartTimeFilterString = fmt.Sprintf(`"%s"."startTime" BETWEEN ? and ?`, dbTransactionInfoCollection)

	transOrderMap = map[apiPb.SortTransactionList]string{
		apiPb.SortTransactionList_SORT_TRANSACTION_LIST_UNSPECIFIED: fmt.Sprintf(`"%s"."startTime"`, dbTransactionInfoCollection),
		apiPb.SortTransactionList_DURATION:                          fmt.Sprintf(`"%s"."endTime" - "%s"."startTime"`, dbTransactionInfoCollection, dbTransactionInfoCollection),
		apiPb.SortTransactionList_BY_TRANSACTION_START_TIME:         fmt.Sprintf(`"%s"."startTime"`, dbTransactionInfoCollection),
		apiPb.SortTransactionList_BY_TRANSACTION_END_TIME:           fmt.Sprintf(`"%s"."endTime"`, dbTransactionInfoCollection),
	}
	groupMap = map[apiPb.GroupTransaction]string{
		apiPb.GroupTransaction_GROUP_TRANSACTION_UNSPECIFIED: transTransactionTypeStr,
		apiPb.GroupTransaction_BY_TYPE:                       transTransactionTypeStr,
		apiPb.GroupTransaction_BY_NAME:                       transNameStr,
		apiPb.GroupTransaction_BY_METHOD:                     transMetaMethodStr,
		apiPb.GroupTransaction_BY_HOST:                       transMetaHostStr,
		apiPb.GroupTransaction_BY_PATH:                       transMetaPathStr,
	}
)

//func (p *Postgres) InsertTransactionInfo(data *apiPb.TransactionInfo) error {
//	info, err := convertToTransactionInfo(data)
//	if err != nil {
//		return err
//	}
//	if err := p.Db.Table(dbTransactionInfoCollection).Create(info).Error; err != nil {
//		return errorDataBase
//	}
//	return nil
//}
//
//func (p *Postgres) GetTransactionInfo(request *apiPb.GetTransactionsRequest) ([]*apiPb.TransactionInfo, int64, error) {
//	timeFrom, timeTo, err := getTimeInt64(request.TimeRange)
//	if err != nil {
//		return nil, -1, err
//	}
//
//	var count int64
//	err = p.Db.Table(dbTransactionInfoCollection).
//		Where(applicationIdFilterString, request.GetApplicationId()).
//		Where(applicationStartTimeFilterString, timeFrom, timeTo).
//		Where(getTransactionsByString(transMetaHostStr, request.GetHost())).
//		Where(getTransactionsByString(transNameStr, request.GetName())).
//		Where(getTransactionsByString(transMetaPathStr, request.GetPath())).
//		Where(getTransactionsByString(transMetaMethodStr, request.GetMethod())).
//		Where(getTransactionTypeWhere(request.GetType())).
//		Where(getTransactionStatusWhere(request.GetStatus())).
//		Count(&count).Error
//	if err != nil {
//		return nil, -1, err
//	}
//
//	offset, limit := getOffsetAndLimit(count, request.GetPagination())
//
//	//TODO: order
//	var statRequests []*TransactionInfo
//	err = p.Db.Table(dbTransactionInfoCollection).
//		Where(applicationIdFilterString, request.GetApplicationId()).
//		Where(applicationStartTimeFilterString, timeFrom, timeTo).
//		Where(getTransactionsByString(transMetaHostStr, request.GetHost())).
//		Where(getTransactionsByString(transNameStr, request.GetName())).
//		Where(getTransactionsByString(transMetaPathStr, request.GetPath())).
//		Where(getTransactionsByString(transMetaMethodStr, request.GetMethod())).
//		Where(getTransactionTypeWhere(request.GetType())).
//		Where(getTransactionStatusWhere(request.GetStatus())).
//		Order(getTransactionOrder(request.GetSort()) + getTransactionDirection(request.GetSort())). //TODO
//		Offset(offset).
//		Limit(limit).
//		Find(&statRequests).
//		Error
//
//	if err != nil {
//		return nil, -1, errorDataBase
//	}
//
//	return convertFromTransactions(statRequests), count, nil
//}
//
//func (p *Postgres) GetTransactionByID(request *apiPb.GetTransactionByIdRequest) (*apiPb.TransactionInfo, []*apiPb.TransactionInfo, error) {
//	var transaction TransactionInfo
//	err := p.Db.Table(dbTransactionInfoCollection).
//		Where(fmt.Sprintf(`"%s"."transactionId" = ?`, dbTransactionInfoCollection), request.GetTransactionId()).
//		First(&transaction).
//		Error
//	if err != nil || &transaction == nil {
//		return nil, nil, err
//	}
//
//	children, err := p.GetTransactionChildren(transaction.TransactionId, "")
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return convertFromTransaction(&transaction), convertFromTransactions(children), nil
//}
//
////passedString is used in order to prevent cycles
//func (p *Postgres) GetTransactionChildren(transactionId, passedString string) ([]*TransactionInfo, error) {
//	if strings.Contains(passedString, transactionId) {
//		return nil, nil
//	}
//
//	var childrenTransactions []*TransactionInfo
//	err := p.Db.Table(dbTransactionInfoCollection).
//		Where(fmt.Sprintf(`"%s"."parentId" = ?`, dbTransactionInfoCollection), transactionId).
//		Find(&childrenTransactions).
//		Error
//	if err != nil {
//		return nil, errorDataBase
//	}
//
//	res := childrenTransactions
//	for _, v := range childrenTransactions {
//		subchildren, err := p.GetTransactionChildren(v.TransactionId, passedString+" "+v.ParentId)
//		if err != nil {
//			return nil, errorDataBase
//		}
//		res = append(res, subchildren...)
//	}
//
//	return res, nil
//}
//
//func (p *Postgres) GetTransactionGroup(request *apiPb.GetTransactionGroupRequest) (map[string]*apiPb.TransactionGroup, error) {
//	timeFrom, timeTo, err := getTimeInt64(request.TimeRange)
//	if err != nil {
//		return nil, err
//	}
//
//	selectString := fmt.Sprintf(
//		`%s as "groupName", COUNT(%s) as "count", COUNT(CASE WHEN "%s"."transactionStatus" = '1' THEN 1 ELSE NULL END) as "successCount", AVG("%s"."endTime"-"%s"."startTime") as "latency", min("%s"."endTime"-"%s"."startTime") as "minTime", max("%s"."endTime"-"%s"."startTime") as "maxTime", min("%s"."endTime") as "lowTime"`,
//		getTransactionsGroupBy(request.GetGroupType()),
//		getTransactionsGroupBy(request.GetGroupType()),
//		dbTransactionInfoCollection,
//		dbTransactionInfoCollection,
//		dbTransactionInfoCollection,
//		dbTransactionInfoCollection,
//		dbTransactionInfoCollection,
//		dbTransactionInfoCollection,
//		dbTransactionInfoCollection,
//		dbTransactionInfoCollection,
//	)
//
//	//TODO: order
//	var groupResult []*GroupResult
//	err = p.Db.Table(dbTransactionInfoCollection).
//		Select(selectString).
//		Where(applicationIdFilterString, request.GetApplicationId()).
//		Where(applicationStartTimeFilterString, timeFrom, timeTo).
//		Where(getTransactionTypeWhere(request.GetType())).
//		Where(getTransactionStatusWhere(request.GetStatus())).
//		Group(getTransactionsGroupBy(request.GetGroupType())).
//		Find(&groupResult).
//		Error
//	if err != nil {
//		return nil, errorDataBase
//	}
//
//	return convertFromGroupResult(groupResult, timeTo), nil
//}

func getTransactionOrder(request *apiPb.SortingTransactionList) string {
	if request == nil {
		return fmt.Sprintf(`"%s"."startTime"`, dbTransactionInfoCollection)
	}
	if res, ok := transOrderMap[request.GetSortBy()]; ok {
		return res
	}
	return fmt.Sprintf(`"%s"."startTime"`, dbTransactionInfoCollection)
}

func getTransactionDirection(request *apiPb.SortingTransactionList) string {
	if request == nil {
		return ` desc`
	}
	if res, ok := directionMap[request.GetDirection()]; ok {
		return res
	}
	return ` desc`
}

func getTransactionsByString(key string, value *wrappers.StringValue) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf(`"%s"."%s" = '%s'`, dbTransactionInfoCollection, key, value.GetValue())
}

func getTransactionTypeWhere(transType apiPb.TransactionType) string {
	if transType == apiPb.TransactionType_TRANSACTION_TYPE_UNSPECIFIED {
		return ""
	}
	return fmt.Sprintf(`"%s"."%s" = '%d'`, dbTransactionInfoCollection, transTransactionTypeStr, transType)
}

func getTransactionStatusWhere(transType apiPb.TransactionStatus) string {
	if transType == apiPb.TransactionStatus_TRANSACTION_CODE_UNSPECIFIED {
		return ""
	}
	return fmt.Sprintf(`"%s"."transactionStatus" = '%d'`, dbTransactionInfoCollection, transType)
}

func getTransactionsGroupBy(group apiPb.GroupTransaction) string {
	if val, ok := groupMap[group]; ok {
		return fmt.Sprintf(`"%s"."%s"`, dbTransactionInfoCollection, val)
	}
	return fmt.Sprintf(`"%s"."%s"`, dbTransactionInfoCollection, transTransactionTypeStr)
}
