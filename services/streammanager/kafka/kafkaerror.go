package kafka

var abortableErrors []string = []string{errNotConnected, errInsufficientData, errInvalidMessage, errUnknownTopicOrPartition,
	errInvalidMessageSize, errNotLeaderForPartition, errRequestTimedOut, errBrokerNotAvailable, errMessageSizeTooLarge, errOffsetMetadataTooLarge,
	errInvalidTopic, errNotEnoughReplicas, errNotEnoughReplicasAfterAppend, errInvalidSessionTimeout, errInvalidGroupID, errUnknownMemberID,
	errTopicAuthorizationFailed, errGroupAuthorizationFailed, errClusterAuthorizationFailed, errInvalidConfig, errInvalidRequest,
}

const (
	errClosedClient                       string = "kafka: tried to use a client that was closed"
	errIncompleteResponse                 string = "kafka: response did not contain all the expected topic/partition blocks"
	errInvalidPartition                   string = "kafka: partitioner returned an invalid partition index"
	errAlreadyConnected                   string = "kafka: broker connection already initiated"
	errNotConnected                       string = "kafka: broker not connected"
	errInsufficientData                   string = "kafka: insufficient data to decode packet, more bytes expected"
	errShuttingDown                       string = "kafka: message received by producer in process of shutting down"
	errUnknown                            string = "kafka server: Unexpected (unknown?) server error."
	errOffsetOutOfRange                   string = "kafka server: The requested offset is outside the range of offsets maintained by the server for the given topic/partition."
	errInvalidMessage                     string = "kafka server: Message contents does not match its CRC."
	errUnknownTopicOrPartition            string = "kafka server: Request was for a topic or partition that does not exist on this broker."
	errInvalidMessageSize                 string = "kafka server: The message has a negative size."
	errLeaderNotAvailable                 string = "kafka server: In the middle of a leadership election, there is currently no leader for this partition and hence it is unavailable for writes."
	errNotLeaderForPartition              string = "kafka server: Tried to send a message to a replica that is not the leader for some partition. Your metadata is out of date."
	errRequestTimedOut                    string = "kafka server: Request exceeded the user-specified time limit in the request."
	errBrokerNotAvailable                 string = "kafka server: Broker not available. Not a client facing error, we should never receive this!!!"
	errReplicaNotAvailable                string = "kafka server: Replica information not available, one or more brokers are down."
	errMessageSizeTooLarge                string = "kafka server: Message was too large, server rejected it to avoid allocation error."
	errStaleControllerEpochCode           string = "kafka server: StaleControllerEpochCode (internal error code for broker-to-broker communication)."
	errOffsetMetadataTooLarge             string = "kafka server: Specified a string larger than the configured maximum for offset metadata."
	errNetworkException                   string = "kafka server: The server disconnected before a response was received."
	errOffsetsLoadInProgress              string = "kafka server: The broker is still loading offsets after a leader change for that offset's topic partition."
	errInvalidTopic                       string = "kafka server: The request attempted to perform an operation on an invalid topic."
	errNotEnoughReplicas                  string = "kafka server: Messages are rejected since there are fewer in-sync replicas than required."
	errNotEnoughReplicasAfterAppend       string = "kafka server: Messages are written to the log, but to fewer in-sync replicas than required."
	errInvalidRequiredAcks                string = "kafka server: The number of required acks is invalid (should be either -1, 0, or 1)."
	errInvalidGroupID                     string = "kafka server: The provided group id was empty."
	errUnknownMemberID                    string = "kafka server: The provided member is not known in the current generation."
	errInvalidSessionTimeout              string = "kafka server: The provided session timeout is outside the allowed range."
	errRebalanceInProgress                string = "kafka server: A rebalance for the group is in progress. Please re-join the group."
	errTopicAuthorizationFailed           string = "kafka server: The client is not authorized to access this topic."
	errGroupAuthorizationFailed           string = "kafka server: The client is not authorized to access this group."
	errClusterAuthorizationFailed         string = "kafka server: The client is not authorized to send this request type."
	errInvalidTimestamp                   string = "kafka server: The timestamp of the message is out of acceptable range."
	errInvalidPartitions                  string = "kafka server: Number of partitions is invalid."
	errInvalidReplicationFactor           string = "kafka server: Replication-factor is invalid."
	errInvalidReplicaAssignment           string = "kafka server: Replica assignment is invalid."
	errInvalidConfig                      string = "kafka server: Configuration is invalid."
	errInvalidRequest                     string = "kafka server: This most likely occurs because of a request being malformed by the client library or the message was sent to an incompatible broker. See the broker logs for more details."
	errUnsupportedForMessageFormat        string = "kafka server: The requested operation is not supported by the message format version."
	errPolicyViolation                    string = "kafka server: Request parameters do not satisfy the configured policy."
	errInvalidProducerEpoch               string = "kafka server: Producer attempted an operation with an old epoch."
	errInvalidTxnState                    string = "kafka server: The producer attempted a transactional operation in an invalid state."
	errInvalidProducerIDMapping           string = "kafka server: The producer attempted to use a producer id which is not currently assigned to its transactional id."
	errInvalidTransactionTimeout          string = "kafka server: The transaction timeout is larger than the maximum value allowed by the broker (as configured by max.transaction.timeout.ms)."
	errConcurrentTransactions             string = "kafka server: The producer attempted to update a transaction while another concurrent operation on the same transaction was ongoing."
	errTransactionCoordinatorFenced       string = "kafka server: The transaction coordinator sending a WriteTxnMarker is no longer the current coordinator for a given producer."
	errTransactionalIDAuthorizationFailed string = "kafka server: Transactional ID authorization failed."
	errSecurityDisabled                   string = "kafka server: Security features are disabled."
	errOperationNotAttempted              string = "kafka server: The broker did not attempt to execute this operation."
	errKafkaStorageError                  string = "kafka server: Disk error when trying to access log file on the disk."
	errLogDirNotFound                     string = "kafka server: The specified log directory is not found in the broker config."
	errUnknownProducerID                  string = "kafka server: The broker could not locate the producer metadata associated with the Producer ID."
	errReassignmentInProgress             string = "kafka server: A partition reassignment is in progress."
	errDelegationTokenAuthDisabled        string = "kafka server: Delegation Token feature is not enabled."
	errDelegationTokenNotFound            string = "kafka server: Delegation Token is not found on server."
	errDelegationTokenOwnerMismatch       string = "kafka server: Specified Principal is not valid Owner/Renewer."
	errDelegationTokenRequestNotAllowed   string = "kafka server: Delegation Token requests are not allowed on PLAINTEXT/1-way SSL channels and on delegation token authenticated channels."
	errDelegationTokenAuthorizationFailed string = "kafka server: Delegation Token authorization failed."
	errDelegationTokenExpired             string = "kafka server: Delegation Token is expired."
	errInvalidPrincipalType               string = "kafka server: Supplied principalType is not supported."
	errNonEmptyGroup                      string = "kafka server: The group is not empty."
	errGroupIDNotFound                    string = "kafka server: The group id does not exist."
	errListenerNotFound                   string = "kafka server: There is no listener on the leader broker that matches the listener on which metadata request was processed."
)
