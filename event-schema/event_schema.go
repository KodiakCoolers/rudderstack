//Event schemas uses countish algorithm by https://github.com/shanemhansen/countish

/*
 *
Table: event_models

| id  | uuid   | write_key | event_type | event_model_identifier | created_at        |
| --- | ------ | --------- | ---------- | ---------------------- | ----------------- |
| 1   | uuid-1 | ksuid-1   | track      | logged_in              | 01, Jan 12: 00 PM |
| 2   | uuid-2 | ksuid-1   | track      | signed_up              | 01, Jan 12: 00 PM |
| 3   | uuid-3 | ksuid-1   | page       | Home Page              | 01, Jan 12: 00 PM |
| 4   | uuid-4 | ksuid-2   | identify   |                        | 01, Jan 12: 00 PM |


Table: schema_versions

| id  | uuid   | event_model_id | schema_hash | schema                          | metadata | first_seen        | last_seen          |
| --- | ------ | -------------- | ----------- | ------------------------------- | -------- | ----------------- | ------------------ |
| 1   | uuid-9 | uuid-1         | hash-1      | {"a": "string", "b": "float64"} | {}       | 01, Jan 12: 00 PM | 01, June 12: 00 PM |
| 2   | uuid-8 | uuid-2         | hash-2      | {"a": "string", "b": "string"}  | {}       | 01, Jan 12: 00 PM | 01, June 12: 00 PM |
| 3   | uuid-7 | uuid-3         | hash-3      | {"a": "string", "c": "float64"} | {}       | 01, Jan 12: 00 PM | 01, June 12: 00 PM |
| 4   | uuid-6 | uuid-2         | hash-1      | {"a": "string", "b": "float64"} | {}       | 01, Jan 12: 00 PM | 01, June 12: 00 PM |

*/

package event_schema

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jeremywohl/flatten"
	"github.com/lib/pq"

	"github.com/rudderlabs/rudder-server/config"
	"github.com/rudderlabs/rudder-server/jobsdb"
	"github.com/rudderlabs/rudder-server/rruntime"
	"github.com/rudderlabs/rudder-server/services/stats"
	"github.com/rudderlabs/rudder-server/utils/logger"
	"github.com/rudderlabs/rudder-server/utils/misc"
	"github.com/rudderlabs/rudder-server/utils/timeutil"
	uuid "github.com/satori/go.uuid"
)

// EventModelT is a struct that represents EVENT_MODELS_TABLE
type EventModelT struct {
	ID              int
	UUID            string `json:"EventID"`
	WriteKey        string `json:"WriteKey"`
	EventType       string `json:"EventType"`
	EventIdentifier string `json:"EventIdentifier"`
	CreatedAt       time.Time
	Schema          json.RawMessage
	Metadata        json.RawMessage `json:"-"`
	PrivateData     json.RawMessage `json:"-"`
	LastSeen        time.Time
	reservoirSample *ReservoirSample
	TotalCount      int64
}

// SchemaVersionT is a struct that represents SCHEMA_VERSIONS_TABLE
type SchemaVersionT struct {
	ID              int64
	UUID            string `json:"VersionID"`
	SchemaHash      string `json:"-"`
	EventModelID    string
	Schema          json.RawMessage
	Metadata        json.RawMessage `json:"-"`
	PrivateData     json.RawMessage `json:"-"`
	FirstSeen       time.Time
	LastSeen        time.Time
	reservoirSample *ReservoirSample
	TotalCount      int64
}

type MetaDataT struct {
	SampledEvents []interface{}
	TotalCount    int64
	Counters      map[string][]*CounterItem `json:"FrequentValues"`
}

type PrivateDataT struct {
	FrequencyCounters []*FrequencyCounter
}

type WriteKey string
type EventType string
type EventIdentifier string

//EventModelMapT : <writeKey, eventType, eventIdentifier> to EventModel Mapping
type EventModelMapT map[WriteKey]map[EventType]map[EventIdentifier]*EventModelT

//SchemaVersionMapT : <event_model_id, schema_hash> to SchemaVersion Mapping
type SchemaVersionMapT map[string]map[string]*SchemaVersionT

// EventSchemaManagerT handles all event-schemas related features
type EventSchemaManagerT struct {
	dbHandle             *sql.DB
	eventModelMap        EventModelMapT
	schemaVersionMap     SchemaVersionMapT
	eventModelLock       sync.RWMutex
	schemaVersionLock    sync.RWMutex
	disableInMemoryCache bool
}

type OffloadedModelT struct {
	UUID            string
	LastSeen        time.Time
	WriteKey        string
	EventType       string
	EventIdentifier string
}

var (
	flushIntervalInS                int
	adminUser                       string
	adminPassword                   string
	reservoirSampleSize             int
	eventSchemaChannel              chan *GatewayEventBatchT
	updatedEventModels              map[string]*EventModelT
	updatedSchemaVersions           map[string]*SchemaVersionT
	offloadedEventModels            map[string]map[string]*OffloadedModelT
	toDeleteEventModelIDs           []string
	toDeleteSchemaVersionIDs        []string
	pkgLogger                       logger.LoggerI
	noOfWorkers                     int
	shouldCaptureNilAsUnknowns      bool
	eventModelLimit                 int
	schemaVersionPerEventModelLimit int
	offloadLoopIntervalInS          int
	offloadThresholdInS             int
)

const EVENT_MODELS_TABLE = "event_models"
const SCHEMA_VERSIONS_TABLE = "schema_versions"

//GatewayEventBatchT : Type sent from gateway
type GatewayEventBatchT struct {
	writeKey   string
	eventBatch string
}

//EventT : Generic type for singular event
type EventT map[string]interface{}

//EventPayloadT : Generic type for gateway event payload
type EventPayloadT struct {
	WriteKey   string
	ReceivedAt string
	Batch      []EventT
}

func loadConfig() {
	adminUser = config.GetEnv("RUDDER_ADMIN_USER", "rudder")
	adminPassword = config.GetEnv("RUDDER_ADMIN_PASSWORD", "rudderstack")
	noOfWorkers = config.GetInt("EventSchemas.noOfWorkers", 128)
	config.RegisterIntConfigVariable(120, &flushIntervalInS, true, 1, "EventSchemas.syncIntervalInS")
	config.RegisterIntConfigVariable(5, &reservoirSampleSize, true, 1, "EventSchemas.sampleEventsSize")
	config.RegisterIntConfigVariable(200, &eventModelLimit, true, 1, "EventSchemas.eventModelLimit")
	config.RegisterIntConfigVariable(20, &schemaVersionPerEventModelLimit, true, 1, "EventSchemas.schemaVersionPerEventModelLimit")
	config.RegisterBoolConfigVariable(false, &shouldCaptureNilAsUnknowns, true, "EventSchemas.captureUnknowns")
	config.RegisterIntConfigVariable(60, &offloadLoopIntervalInS, true, 1, "EventSchemas.offloadLoopIntervalInS")
	config.RegisterIntConfigVariable(1800, &offloadThresholdInS, true, 1, "EventSchemas.offloadThresholdInS")

	if adminPassword == "rudderstack" {
		fmt.Println("[EventSchemas] You are using default password. Please change it by setting env variable RUDDER_ADMIN_PASSWORD")
	}
}

func init() {
	loadConfig()
	pkgLogger = logger.NewLogger().Child("event-schema")
}

//RecordEventSchema : Records event schema for every event in the batch
func (manager *EventSchemaManagerT) RecordEventSchema(writeKey string, eventBatch string) bool {
	select {
	case eventSchemaChannel <- &GatewayEventBatchT{writeKey, eventBatch}:
	default:
		stats.NewTaggedStat("dropped_events_count", stats.CountType, stats.Tags{"module": "event_schemas", "writeKey": writeKey}).Increment()
	}
	return true
}

func (manager *EventSchemaManagerT) updateEventModelCache(eventModel *EventModelT, toCreateOrUpdate bool) {
	eventModelID := eventModel.UUID
	writeKey := eventModel.WriteKey
	eventType := eventModel.EventType
	eventIdentifier := eventModel.EventIdentifier

	_, ok := manager.eventModelMap[WriteKey(writeKey)]
	if !ok {
		manager.eventModelMap[WriteKey(writeKey)] = make(map[EventType]map[EventIdentifier]*EventModelT)
	}
	_, ok = manager.eventModelMap[WriteKey(writeKey)][EventType(eventType)]
	if !ok {
		manager.eventModelMap[WriteKey(writeKey)][EventType(eventType)] = make(map[EventIdentifier]*EventModelT)
	}
	manager.eventModelMap[WriteKey(writeKey)][EventType(eventType)][EventIdentifier(eventIdentifier)] = eventModel

	if toCreateOrUpdate {
		updatedEventModels[eventModelID] = eventModel
	}
}

func (manager *EventSchemaManagerT) deleteFromEventModelCache(eventModel *EventModelT) {
	writeKey := eventModel.WriteKey
	eventType := eventModel.EventType
	eventIdentifier := eventModel.EventIdentifier

	delete(manager.eventModelMap[WriteKey(writeKey)][EventType(eventType)], EventIdentifier(eventIdentifier))
}

func (manager *EventSchemaManagerT) deleteFromSchemaVersionCache(schemaVersion *SchemaVersionT) {
	eventModelID := schemaVersion.EventModelID
	schemaHash := schemaVersion.SchemaHash

	delete(manager.schemaVersionMap[eventModelID], schemaHash)
}

func (manager *EventSchemaManagerT) deleteModelFromSchemaVersionCache(eventModel *EventModelT) {
	delete(manager.schemaVersionMap, eventModel.UUID)
}

func (manager *EventSchemaManagerT) updateSchemaVersionCache(schemaVersion *SchemaVersionT, toCreateOrUpdate bool) {
	eventModelID := schemaVersion.EventModelID
	schemaHash := schemaVersion.SchemaHash

	_, ok := manager.schemaVersionMap[eventModelID]
	if !ok {
		manager.schemaVersionMap[eventModelID] = make(map[string]*SchemaVersionT)
	}
	manager.schemaVersionMap[eventModelID][schemaHash] = schemaVersion

	if toCreateOrUpdate {
		updatedSchemaVersions[schemaVersion.UUID] = schemaVersion
	}
}

/*
 *
| Event Type | event_type | event_model_identfier |
| ---------- | ---------- | --------------------- |
| track      | track      | event["event"]        |
| page       | page       | event["name"]         |
| screen     | screen     | event["name"]         |
| identify   | identify   | ""                    |
| alias      | alias      | ""                    |
| group      | group      | ""                    |
*
* All event types and schema versions are generated by grouping according to the table above.
* Eg:
*    <track, login> will always be of same event_model. Different payloads will result in different schema_versions
*    <track, login> will always be of same event_model. Different payloads will result in different schema_versions
*    <page, home-page> will always be of same event_model. Different payloads will result in different schema_versions
*    <identify> There will be only identify event_model per source. Schema versions can change with different traits.
*
* This function is goroutine-safe. We can scale multiple go-routines calling this function,
* but since this method does mostly in-memory operations and has locks, there might not be much perfomance improvement.
*/
func (manager *EventSchemaManagerT) handleEvent(writeKey string, event EventT) {
	eventType, ok := event["type"].(string)
	if !ok {
		pkgLogger.Debugf("[EventSchemas] Invalid or no eventType")
		return
	}
	eventIdentifier := ""
	if eventType == "track" {
		eventIdentifier, ok = event["event"].(string)
	}
	if !ok {
		pkgLogger.Debugf("[EventSchemas] Invalid event idenitfier")
		return
	}

	//TODO: Create locks on every event_model to improve scaling this
	manager.eventModelLock.Lock()
	manager.schemaVersionLock.Lock()
	defer manager.eventModelLock.Unlock()
	defer manager.schemaVersionLock.Unlock()
	totalEventModels := 0
	for _, v := range manager.eventModelMap[WriteKey(writeKey)] {
		totalEventModels += len(v)
	}
	totalEventModels += len(offloadedEventModels[writeKey])
	eventModel, ok := manager.eventModelMap[WriteKey(writeKey)][EventType(eventType)][EventIdentifier(eventIdentifier)]
	if !ok {
		var wasOffloaded bool
		var offloadedModel *OffloadedModelT
		if x, ok := offloadedEventModels[writeKey]; ok {
			offloadedModel, wasOffloaded = x[eventTypeIdentifier(eventType, eventIdentifier)]
		}
		if wasOffloaded {
			manager.reloadModel(offloadedModel)
			eventModel, ok = manager.eventModelMap[WriteKey(writeKey)][EventType(eventType)][EventIdentifier(eventIdentifier)]
			if !ok {
				pkgLogger.Errorf(`[EventSchemas] Failed to reload event +%v, writeKey: %s, eventType: %s, eventIdentifier: %s`, offloadedModel.UUID, writeKey, eventType, eventIdentifier)
				return
			}
		} else {
			eventModelID := uuid.NewV4().String()
			eventModel = &EventModelT{
				UUID:            eventModelID,
				WriteKey:        writeKey,
				EventType:       eventType,
				EventIdentifier: eventIdentifier,
				Schema:          []byte("{}"),
			}
			eventModel.reservoirSample = NewReservoirSampler(reservoirSampleSize, 0, 0)

			if totalEventModels >= eventModelLimit {
				oldestModel := manager.oldestSeenModel(writeKey)
				toDeleteEventModelIDs = append(toDeleteEventModelIDs, oldestModel.UUID)
				manager.deleteFromEventModelCache(oldestModel)
				manager.deleteModelFromSchemaVersionCache(oldestModel)
			}
			manager.updateEventModelCache(eventModel, true)
		}
	}
	eventModel.LastSeen = timeutil.Now()

	eventMap := map[string]interface{}(event)
	flattenedEvent, err := flatten.Flatten((eventMap), "", flatten.DotStyle)
	if err != nil {
		pkgLogger.Debug(fmt.Sprintf("[EventSchemas] Failed to flatten the event +%v with error: %s", eventMap, err.Error()))
		return
	}

	schema := getSchema(flattenedEvent)
	schemaHash := getSchemaHash(schema)
	computeFrequencies(flattenedEvent, schemaHash)
	computeFrequencies(flattenedEvent, eventModel.UUID)

	schemaVersion, schemaFoundInCache := manager.schemaVersionMap[eventModel.UUID][schemaHash]

	if !schemaFoundInCache {
		versionID := uuid.NewV4().String()
		schemaVersion = manager.NewSchemaVersion(versionID, schema, schemaHash, eventModel.UUID)
		eventModel.mergeSchema(schemaVersion)

		if len(manager.schemaVersionMap[eventModel.UUID]) >= schemaVersionPerEventModelLimit {
			stats.NewTaggedStat("dropped_schema_versions_count", stats.CountType, stats.Tags{"module": "event_schemas", "eventModelID": eventModel.UUID}).Increment()
			oldestVersion := manager.oldestSeenVersion(eventModel.UUID)
			toDeleteSchemaVersionIDs = append(toDeleteSchemaVersionIDs, oldestVersion.UUID)
			manager.deleteFromSchemaVersionCache(oldestVersion)
		}
	}
	schemaVersion.LastSeen = timeutil.Now()
	manager.updateSchemaVersionCache(schemaVersion, true)

	eventModel.reservoirSample.add(event, true)
	schemaVersion.reservoirSample.add(event, true)
	updatedEventModels[eventModel.UUID] = eventModel
}

func (manager *EventSchemaManagerT) oldestSeenModel(writeKey string) *EventModelT {
	oldestSeenModel := &EventModelT{}
	var minLastSeen time.Time
	for _, eventIdentifierMap := range manager.eventModelMap[WriteKey(writeKey)] {
		for _, model := range eventIdentifierMap {
			if !model.LastSeen.IsZero() && (model.LastSeen.Sub(minLastSeen).Seconds() <= 0 || minLastSeen.IsZero()) {
				oldestSeenModel = model
				minLastSeen = model.LastSeen
			}
		}
	}
	for _, offloadedModel := range offloadedEventModels[writeKey] {
		if !offloadedModel.LastSeen.IsZero() && (offloadedModel.LastSeen.Sub(minLastSeen).Seconds() <= 0 || minLastSeen.IsZero()) {
			oldestSeenModel.UUID = offloadedModel.UUID
			oldestSeenModel.WriteKey = offloadedModel.WriteKey
			oldestSeenModel.EventType = offloadedModel.EventType
			oldestSeenModel.EventIdentifier = offloadedModel.EventIdentifier
			minLastSeen = offloadedModel.LastSeen
		}
	}
	return oldestSeenModel
}

func (manager *EventSchemaManagerT) oldestSeenVersion(modelID string) *SchemaVersionT {
	var oldestSeenSchemaVersion *SchemaVersionT
	var minLastSeeen time.Time
	for _, schemaVersion := range manager.schemaVersionMap[modelID] {
		if minLastSeeen.IsZero() {
			oldestSeenSchemaVersion = schemaVersion
			minLastSeeen = schemaVersion.LastSeen
		}
		if schemaVersion.LastSeen.Sub(minLastSeeen).Seconds() <= 0 {
			oldestSeenSchemaVersion = schemaVersion
			minLastSeeen = schemaVersion.LastSeen
		}
	}
	return oldestSeenSchemaVersion
}

func (em *EventModelT) mergeSchema(sv *SchemaVersionT) {
	masterSchema := make(map[string]string)
	err := json.Unmarshal(em.Schema, &masterSchema)
	assertError(err)

	schema := make(map[string]string)
	err = json.Unmarshal(sv.Schema, &schema)
	assertError(err)

	errors := make([]string, 0)
	for k := range schema {
		t, ok := masterSchema[k]
		if !ok {
			masterSchema[k] = schema[k]
			continue
		}
		if !strings.Contains(t, schema[k]) {
			masterSchema[k] = fmt.Sprintf("%s,%s", t, schema[k])
		}
	}

	if len(errors) > 0 {
		pkgLogger.Errorf("EventModel with ID: %s has encountered following disparities:\n%s", em.ID, strings.Join(errors, "\n"))
	}

	masterSchemaJSON, err := json.Marshal(masterSchema)
	assertError(err)
	em.Schema = masterSchemaJSON
}

//NewSchemaVersion should be used when a schemaVersion is not found in its cache and requires, a schemaVersionID for the newSchema and the eventModelID to which it belongs along with schema and schemaHash
func (manager *EventSchemaManagerT) NewSchemaVersion(versionID string, schema map[string]string, schemaHash string, eventModelID string) *SchemaVersionT {
	schemaJSON, err := json.Marshal(schema)
	assertError(err)

	schemaVersion := &SchemaVersionT{
		UUID:         versionID,
		SchemaHash:   schemaHash,
		EventModelID: eventModelID,
		Schema:       schemaJSON,
		FirstSeen:    timeutil.Now(),
		LastSeen:     timeutil.Now(),
	}
	schemaVersion.reservoirSample = NewReservoirSampler(reservoirSampleSize, 0, 0)
	return schemaVersion
}

func (manager *EventSchemaManagerT) recordEvents() {
	for gatewayEventBatch := range eventSchemaChannel {

		var eventPayload EventPayloadT
		err := json.Unmarshal([]byte(gatewayEventBatch.eventBatch), &eventPayload)
		assertError(err)
		for _, event := range eventPayload.Batch {
			manager.handleEvent(eventPayload.WriteKey, event)
		}
	}
}

func getMetadataJSON(reservoirSample *ReservoirSample, schemaHash string) []byte {
	metadata := &MetaDataT{
		SampledEvents: reservoirSample.getSamples(),
		TotalCount:    reservoirSample.getTotalCount(),
	}
	metadata.Counters = getSchemaVersionCounters(schemaHash)

	metadataJSON, err := json.Marshal(metadata)
	pkgLogger.Debugf("[EventSchemas] Metadata JSON: %s", string(metadataJSON))
	assertError(err)
	return metadataJSON
}

func getPrivateDataJSON(schemaHash string) []byte {
	privateData := &PrivateDataT{
		FrequencyCounters: getAllFrequencyCounters(schemaHash),
	}

	privateDataJSON, err := json.Marshal(privateData)
	pkgLogger.Debugf("[EventSchemas] Private Data JSON: %s", string(privateDataJSON))
	assertError(err)
	return privateDataJSON

}

func (manager *EventSchemaManagerT) flushEventSchemas() {
	// This will run forever. If you want to quit in between, change it to ticker and call stop()
	// Otherwise the ticker won't be GC'ed
	ticker := time.Tick(time.Duration(flushIntervalInS) * time.Second)
	for range ticker {

		// If needed, copy the maps and release the lock immediately
		manager.eventModelLock.Lock()
		manager.schemaVersionLock.Lock()

		schemaVersionsInCache := make([]*SchemaVersionT, 0)
		for _, sv := range updatedSchemaVersions {
			schemaVersionsInCache = append(schemaVersionsInCache, sv)
		}

		if len(updatedEventModels) == 0 && len(schemaVersionsInCache) == 0 {
			manager.eventModelLock.Unlock()
			manager.schemaVersionLock.Unlock()
			continue
		}

		txn, err := manager.dbHandle.Begin()
		assertError(err)

		// Handle Event Models
		if len(updatedEventModels) > 0 {
			eventModelIds := make([]string, 0, len(updatedEventModels))
			for _, em := range updatedEventModels {
				eventModelIds = append(eventModelIds, em.UUID)
			}

			eventModelIds = append(eventModelIds, toDeleteEventModelIDs...)

			deleteOldEventModelsSQL := fmt.Sprintf(`DELETE FROM %s WHERE uuid IN ('%s')`, EVENT_MODELS_TABLE, strings.Join(eventModelIds, "', '"))
			_, err := txn.Exec(deleteOldEventModelsSQL)
			assertTxnError(err, txn)

			if len(toDeleteEventModelIDs) > 0 {
				deleteVersionsForDroppedModelsSQL := fmt.Sprintf(`DELETE FROM %s WHERE event_model_id IN ('%s')`, SCHEMA_VERSIONS_TABLE, strings.Join(toDeleteEventModelIDs, "', '"))
				_, err := txn.Exec(deleteVersionsForDroppedModelsSQL)
				assertTxnError(err, txn)
			}

			stmt, err := txn.Prepare(pq.CopyIn(EVENT_MODELS_TABLE, "uuid", "write_key", "event_type", "event_model_identifier", "schema", "metadata", "private_data", "last_seen", "total_count"))
			assertTxnError(err, txn)
			//skipcq: SCC-SA9001
			defer stmt.Close()
			for eventModelID, eventModel := range updatedEventModels {
				metadataJSON := getMetadataJSON(eventModel.reservoirSample, eventModel.UUID)
				privateDataJSON := getPrivateDataJSON(eventModel.UUID)
				eventModel.TotalCount = eventModel.reservoirSample.totalCount

				_, err = stmt.Exec(eventModelID, eventModel.WriteKey, eventModel.EventType, eventModel.EventIdentifier, string(eventModel.Schema), string(metadataJSON), string(privateDataJSON), eventModel.LastSeen, eventModel.TotalCount)
				assertTxnError(err, txn)
			}
			_, err = stmt.Exec()
			assertTxnError(err, txn)
			stats.NewTaggedStat("update_event_model_count", stats.GaugeType, stats.Tags{"module": "event_schemas"}).Gauge(len(eventModelIds))
			pkgLogger.Debugf("[EventSchemas][Flush] %d new event types", len(updatedEventModels))
		}

		//Handle Schema Versions
		if len(schemaVersionsInCache) > 0 {
			versionIDs := make([]string, 0, len(schemaVersionsInCache))
			for _, sv := range schemaVersionsInCache {
				versionIDs = append(versionIDs, sv.UUID)
			}

			versionIDs = append(versionIDs, toDeleteSchemaVersionIDs...)

			deleteOldVersionsSQL := fmt.Sprintf(`DELETE FROM %s WHERE uuid IN ('%s')`, SCHEMA_VERSIONS_TABLE, strings.Join(versionIDs, "', '"))
			_, err := txn.Exec(deleteOldVersionsSQL)
			assertTxnError(err, txn)

			stmt, err := txn.Prepare(pq.CopyIn(SCHEMA_VERSIONS_TABLE, "uuid", "event_model_id", "schema_hash", "schema", "metadata", "private_data", "first_seen", "last_seen", "total_count"))
			assertTxnError(err, txn)
			//skipcq: SCC-SA9001
			defer stmt.Close()
			for _, sv := range schemaVersionsInCache {
				metadataJSON := getMetadataJSON(sv.reservoirSample, sv.SchemaHash)
				privateDataJSON := getPrivateDataJSON(sv.SchemaHash)
				sv.TotalCount = sv.reservoirSample.totalCount

				_, err = stmt.Exec(sv.UUID, sv.EventModelID, sv.SchemaHash, string(sv.Schema), string(metadataJSON), string(privateDataJSON), sv.FirstSeen, sv.LastSeen, sv.TotalCount)
				assertTxnError(err, txn)
			}
			_, err = stmt.Exec()
			assertTxnError(err, txn)
			stats.NewTaggedStat("update_schema_version_count", stats.GaugeType, stats.Tags{"module": "event_schemas"}).Gauge(len(versionIDs))
			pkgLogger.Debugf("[EventSchemas][Flush] %d new schema versions", len(schemaVersionsInCache))
		}

		err = txn.Commit()
		assertTxnError(err, txn)

		updatedEventModels = make(map[string]*EventModelT)
		updatedSchemaVersions = make(map[string]*SchemaVersionT)
		toDeleteEventModelIDs = []string{}

		manager.schemaVersionLock.Unlock()
		manager.eventModelLock.Unlock()
	}
}

func eventTypeIdentifier(eventType, eventIdentifier string) string {
	return fmt.Sprintf(`%s::%s`, eventType, eventIdentifier)
}

func (manager *EventSchemaManagerT) offloadSchemas() {
	for {
		time.Sleep(time.Duration(offloadLoopIntervalInS) * time.Second)
		manager.eventModelLock.Lock()
		manager.schemaVersionLock.Lock()
		for _, modelsByWriteKey := range manager.eventModelMap {
			for _, modelsByEventType := range modelsByWriteKey {
				for _, model := range modelsByEventType {
					if timeutil.Now().Sub(model.LastSeen) > time.Duration(offloadThresholdInS)*time.Second {
						if _, ok := offloadedEventModels[model.WriteKey]; !ok {
							offloadedEventModels[model.WriteKey] = make(map[string]*OffloadedModelT)
						}
						offloadedEventModels[model.WriteKey][eventTypeIdentifier(model.EventType, model.EventIdentifier)] = &OffloadedModelT{UUID: model.UUID, LastSeen: model.LastSeen, WriteKey: model.WriteKey, EventType: model.EventType, EventIdentifier: model.EventIdentifier}
						manager.deleteFromEventModelCache(model)
						manager.deleteModelFromSchemaVersionCache(model)
					}
				}
			}
		}
		manager.schemaVersionLock.Unlock()
		manager.eventModelLock.Unlock()
	}
}

func (manager *EventSchemaManagerT) reloadModel(offloadedModel *OffloadedModelT) {
	pkgLogger.Infof("reloading event model from db: %s\n", offloadedModel.UUID)
	manager.populateEventModels(offloadedModel.UUID)
	manager.populateSchemaVersions(offloadedModel.UUID)
	delete(offloadedEventModels[offloadedModel.WriteKey], eventTypeIdentifier(offloadedModel.EventType, offloadedModel.EventIdentifier))
}

// TODO: Move this into some DB manager
func createDBConnection() *sql.DB {
	psqlInfo := jobsdb.GetConnectionString()
	var err error
	dbHandle, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = dbHandle.Ping()
	if err != nil {
		panic(err)
	}
	return dbHandle
}

func assertError(err error) {
	if err != nil {
		panic(err)
	}
}

func assertTxnError(err error, txn *sql.Tx) {
	if err != nil {
		txn.Rollback()
		pkgLogger.Info(fmt.Sprintf("%#v\n", err))
		pkgLogger.Info(fmt.Sprintf("%#v\n", txn))
		panic(err)
	}
}

func (manager *EventSchemaManagerT) populateEventModels(uuidFilters ...string) {

	var uuidFilter string
	if len(uuidFilters) > 0 {
		uuidFilter = fmt.Sprintf(`WHERE uuid in ('%s')`, strings.Join(uuidFilters, "', '"))
	}

	eventModelsSelectSQL := fmt.Sprintf(`SELECT * FROM %s %s`, EVENT_MODELS_TABLE, uuidFilter)

	rows, err := manager.dbHandle.Query(eventModelsSelectSQL)
	assertError(err)
	defer rows.Close()

	for rows.Next() {
		var eventModel EventModelT
		err := rows.Scan(&eventModel.ID, &eventModel.UUID, &eventModel.WriteKey, &eventModel.EventType,
			&eventModel.EventIdentifier, &eventModel.CreatedAt, &eventModel.Schema, &eventModel.Metadata,
			&eventModel.PrivateData, &eventModel.TotalCount, &eventModel.LastSeen)

		assertError(err)

		var metadata MetaDataT
		err = json.Unmarshal(eventModel.Metadata, &metadata)
		assertError(err)

		var privateData PrivateDataT
		err = json.Unmarshal(eventModel.PrivateData, &privateData)
		assertError(err)

		eventModel.reservoirSample = NewReservoirSampler(reservoirSampleSize, len(metadata.SampledEvents), metadata.TotalCount)
		for sampledEvent := range metadata.SampledEvents {
			eventModel.reservoirSample.add(sampledEvent, false)
		}
		manager.updateEventModelCache(&eventModel, false)
	}
}

func (manager *EventSchemaManagerT) populateSchemaVersions(modelIDFilters ...string) {
	var modelIDFilter string
	if len(modelIDFilters) > 0 {
		modelIDFilter = fmt.Sprintf(`WHERE event_model_id in ('%s')`, strings.Join(modelIDFilters, "', '"))
	}

	schemaVersionsSelectSQL := fmt.Sprintf(`SELECT id, uuid, event_model_id, schema_hash, schema, metadata, private_data,first_seen, last_seen, total_count FROM %s %s`, SCHEMA_VERSIONS_TABLE, modelIDFilter)

	rows, err := manager.dbHandle.Query(schemaVersionsSelectSQL)
	assertError(err)
	defer rows.Close()

	for rows.Next() {
		var schemaVersion SchemaVersionT
		err := rows.Scan(&schemaVersion.ID, &schemaVersion.UUID, &schemaVersion.EventModelID, &schemaVersion.SchemaHash,
			&schemaVersion.Schema, &schemaVersion.Metadata, &schemaVersion.PrivateData, &schemaVersion.FirstSeen, &schemaVersion.LastSeen, &schemaVersion.TotalCount)
		assertError(err)

		var metadata MetaDataT
		err = json.Unmarshal(schemaVersion.Metadata, &metadata)
		assertError(err)

		var privateData PrivateDataT
		err = json.Unmarshal(schemaVersion.PrivateData, &privateData)
		assertError(err)

		schemaVersion.reservoirSample = NewReservoirSampler(reservoirSampleSize, len(metadata.SampledEvents), metadata.TotalCount)
		for sampledEvent := range metadata.SampledEvents {
			schemaVersion.reservoirSample.add(sampledEvent, false)
		}

		manager.updateSchemaVersionCache(&schemaVersion, false)

		populateFrequencyCounters(schemaVersion.SchemaHash, privateData.FrequencyCounters)
	}
}

// This should be called during the Initialize() to populate existing event Schemas
func (manager *EventSchemaManagerT) populateEventSchemas() {
	pkgLogger.Infof(`Populating event models and their schema versions into in-memory`)
	manager.populateEventModels()
	manager.populateSchemaVersions()
}

func getSchema(flattenedEvent map[string]interface{}) map[string]string {
	schema := make(map[string]string)
	for k, v := range flattenedEvent {
		reflectType := reflect.TypeOf(v)
		if reflectType != nil {
			schema[k] = reflectType.String()
		} else {
			if !(v == nil && !shouldCaptureNilAsUnknowns) {
				schema[k] = "unknown"
				pkgLogger.Errorf("[EventSchemas] Got invalid reflectType %+v", v)
			}
		}
	}
	return schema
}

func getSchemaHash(schema map[string]string) string {
	keys := make([]string, 0, len(schema))
	for k := range schema {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(schema[k])
		sb.WriteString(",")
	}

	schemaHash := misc.GetMD5Hash(sb.String())
	return schemaHash
}

func computeFrequencies(flattenedEvent map[string]interface{}, schemaHash string) {
	// Frequency Counting: Second pass, dependent on schemaHash
	for k, v := range flattenedEvent {
		fc := getFrequencyCounter(schemaHash, k)
		stringVal := fmt.Sprintf("%v", v)
		fc.Observe(stringVal)
	}
}

func (manager *EventSchemaManagerT) Setup() {
	pkgLogger.Info("[EventSchemas] Setting up eventSchemas...")
	// Clean this up
	manager.dbHandle = createDBConnection()

	// Following data structures store events and schemas since last flush
	updatedEventModels = make(map[string]*EventModelT)
	updatedSchemaVersions = make(map[string]*SchemaVersionT)

	manager.eventModelMap = make(EventModelMapT)
	manager.schemaVersionMap = make(SchemaVersionMapT)

	offloadedEventModels = make(map[string]map[string]*OffloadedModelT)

	if !manager.disableInMemoryCache {
		manager.populateEventSchemas()
	}
	eventSchemaChannel = make(chan *GatewayEventBatchT, 10000)

	for i := 0; i < noOfWorkers; i++ {
		rruntime.Go(func() {
			manager.recordEvents()
		})
	}

	rruntime.Go(func() {
		manager.flushEventSchemas()
	})

	rruntime.Go(func() {
		manager.offloadSchemas()
	})

	pkgLogger.Info("[EventSchemas] Set up eventSchemas successful.")
}
