// MetaLocker is a data distribution framework that enables secure and privacy preserving data exchange between participants of a knowledge network.
//
// Source: https://github.com/piprate/metalocker/ledger/onflow/contracts
//
access(all)
contract MetaLocker {

    // Events
    //
    access(all)
    event MetaLockerInitialized()
    access(all)
    event BlockOpened(chain: Address, id: UInt64, parentHash: String)
    access(all)
    event BlockConfirmed(chain: Address, id: UInt64, parentHash: String, nonce: String, recordCount: UInt32, hash: String)
    access(all)
    event RecordPublished(id: String, blockID: UInt64)
    access(all)
    event RecordRevoked(id: String, blockID: UInt64)

    access(self)
    var blocks: {UInt64: MetaBlock}

    access(self)
    var records: {String: Record}

    access(self)
    var dataAssetCounters: {String: UInt64}

    access(self)
    var assetHeads: {String: String}

    access(self)
    var prevBlockHash: [UInt8]
    access(self)
    var currentBlockBody: [UInt8]
    access(self)
    var currentRecordCount: UInt32
    access(self)
    var currentBlockNumber: UInt64
    access(self)
    var genesisBlockHeight: UInt64
    access(self)
    var currentBlockHeight: UInt64

    // Record represents one data transaction (lease, revocation, etc).
    // It contains no details which would allow a third party observer to identify
    // the participants or the nature of this transaction.
    //
    access(all)
    struct Record {
        // ID of the record. Currently it's a hash of the record generated
        // by the Seal function (see below).
        access(all)
        let id: String

        // RoutingKey is a public key from the locker HD structure. It can be
        // used to filter specific messages from the ledger.
        access(all)
        let routingKey: String

        // Index of the HD key used to produce the routing key.
        access(all)
        let keyIndex: UInt32

        // Type of the operation. May be removed from the record in the future.
        access(all)
        let operationType: UInt32

        // OperationAddress is the address of the Operation (can be an asset ID, IPFS address, etc.)
        access(all)
        let address: String

        // Flags contain a set of flags that modify the record's behaviour.
        // See RecordFlagXXX constants for examples.
        access(all)
        let flags: UInt32

        // AuthorisingCommitment is binary data which allows the originator
        // of the transaction to prove their role, without disclosing any
        // other information about this transaction
        access(all)
        let ac: [UInt8]

        // AuthorisingCommitmentType. for future use: there may be different
        // types of commitment structures.
        access(all)
        let acType: UInt8

        // RequestingCommitment is binary data which allows the recipient of
        // the transaction to prove their right to access data without disclosing
        // any other information about this transaction
        access(all)
        let rc: [UInt8]

        // RequestingCommitmentType. for future use: there may be different types
        // of commitment structures.
        access(all)
        let rcType: UInt8

        // ImpressionCommitment is binary data which allows a party to prove
        // that this record contains a specific impression, by combining
        // the impression ID with another artifact, a trapdoor.
        access(all)
        let ic: [UInt8]

        // ImpressionCommitmentType. for future use: there may be different
        // types of commitment structures.
        access(all)
        let icType: UInt8

        // DataAssets is a list of data assets (blobs) attached to the record
        access(all)
        let dataAssets: [String]

        // Lease Revocation fields

        access(all)
        let subjectRecord: String
        access(all)
        let revocationProof: [[UInt8]]

        // Head fields

        // HeadID is unique ID of the asset head.
        access(all)
        let headID: String
        // HeadBody contains base64-encoded, encrypted head body.
        access(all)
        let headBody: String

        // Signature contains a digital signature of the record, signed by
        // the record's private HD key
        access(all)
        let signature: String

        access(all)
        var status: String
        access(all)
        var blockNumber: UInt64

        init(id: String, routingKey: String, keyIndex: UInt32, operationType: UInt32, address: String,
                flags: UInt32, ac: [UInt8], acType: UInt8, rc: [UInt8], rcType: UInt8, ic: [UInt8],
                icType: UInt8, dataAssets: [String], subjectRecord: String, revocationProof: [[UInt8]],
                headID: String, headBody: String, signature: String, status: String, blockNumber: UInt64) {
            self.id = id
            self.routingKey = routingKey
            self.keyIndex = keyIndex
            self.operationType = operationType
            self.address = address
            self.flags = flags
            self.ac = ac
            self.acType = acType
            self.rc = rc
            self.rcType = rcType
            self.ic = ic
            self.icType = icType
            self.dataAssets = dataAssets
            self.subjectRecord = subjectRecord
            self.revocationProof = revocationProof
            self.headID = headID
            self.headBody = headBody
            self.signature = signature
            self.status = status
            self.blockNumber = 0
        }

        access(contract)
        fun setStatus(val: String) {
            self.status = val
        }

        access(contract)
        fun setBlockNumber(val: UInt64) {
            self.blockNumber = val
        }
    }

    access(all)
    struct MetaBlock {
        access(all)
        let number: UInt64
        access(all)
        var nonce: String
        access(all)
        var hash: String
        access(all)
        var parentHash: String
        access(all)
        var status: UInt8
        access(all)
        var records: [[String]]

        init(number: UInt64, parentHash: String, status: UInt8) {
            self.number = number
            self.nonce = ""
            self.hash = ""
            self.parentHash = parentHash
            self.status = status
            self.records = []
        }

        access(contract)
        fun setStatus(val: UInt8) {
            self.status = val
        }

        access(contract)
        fun confirm(nonce: String, hash: String) {
            self.nonce = nonce
            self.hash = hash
            self.status = 2
        }

        access(contract)
        fun addRecord(id: String, routingKey: String, keyIndex: String) {
            self.records.append([id, routingKey, keyIndex])
        }
    }

    access(contract)
    fun rollBlock(blockHeight: UInt64) {
        let nonce = revertibleRandom<UInt32>().toBigEndianBytes()
        let currentRecordCount = UInt32(self.records.length)  // FIXME
        let base = self.prevBlockHash
            .concat(nonce)
            .concat(currentRecordCount.toBigEndianBytes())
            .concat(self.currentBlockBody)
        let hash = HashAlgorithm.SHA2_256.hash(base)

        self.blocks[self.currentBlockNumber]!.confirm(nonce: String.encodeHex(nonce), hash: String.encodeHex(hash))

        let block = self.blocks[self.currentBlockNumber]!

        emit BlockConfirmed(
            chain: self.account.address,
            id: self.currentBlockNumber,
            parentHash: block.parentHash,
            nonce: block.nonce,
            recordCount: UInt32(block.records.length),
            hash: block.hash)

        self.currentRecordCount = 0
        self.currentBlockBody = []
        self.currentBlockHeight = blockHeight
        self.currentBlockNumber = self.currentBlockNumber + UInt64(1)
        self.prevBlockHash = hash

        self.blocks[self.currentBlockNumber] = MetaBlock(number: self.currentBlockNumber, parentHash: String.encodeHex(self.prevBlockHash), status: 1)
        emit BlockOpened(
            chain: self.account.address,
            id: self.currentBlockNumber,
            parentHash: String.encodeHex(self.prevBlockHash))
    }

    access(all)
    fun submitRecord(record: Record) {
        pre {
            record.id != "" : "empty record ID"
            record.signature != "" : "unsigned record"
        }

        if self.records.containsKey(record.id) {
            panic("record already exists")
        }

        let newBlockHeight = getCurrentBlock().height

        if newBlockHeight > self.currentBlockHeight {
            self.rollBlock(blockHeight: newBlockHeight)
        }

        switch record.operationType {
        case UInt32(1): // lease
            self.submitLease(record: record, blockNumber: self.currentBlockNumber)
        case UInt32(2): // lease revocation
            self.submitLeaseRevocation(record: record, blockNumber: self.currentBlockNumber)
        case UInt32(3): // set asset head
            self.submitAssetHead(record: record, blockNumber: self.currentBlockNumber)
        default:
            panic("unsupported operation type: ".concat(record.operationType.toString()))
        }

        self.blocks[self.currentBlockNumber]!.addRecord(id: record.id, routingKey: record.routingKey, keyIndex: record.keyIndex.toString())
        self.currentBlockBody = self.currentBlockBody.concat(record.id.utf8)
        emit RecordPublished(id: record.id, blockID: self.currentBlockNumber)
    }

    access(contract)
    fun submitLease(record: Record, blockNumber: UInt64) {
        //log(record)

        record.setStatus(val: "published")
        record.setBlockNumber(val: blockNumber)
        self.records[record.id] = record

        // update data asset counters

        let incrementCounter = fun (dataAssetID: String) {
            var counter = self.dataAssetCounters[dataAssetID]
            if counter != nil {
                self.dataAssetCounters[dataAssetID] = counter! + 1
            } else {
                self.dataAssetCounters[dataAssetID] = 1
            }
        }

        for id in record.dataAssets {
            incrementCounter(id)
        }
        incrementCounter(record.address)
    }

    access(contract)
    fun submitLeaseRevocation(record: Record, blockNumber: UInt64) {
        let subjRec = self.records[record.subjectRecord]
            ?? panic("Subject record not found: ".concat(record.subjectRecord))

        // check revocation proof

        assert(record.revocationProof.length > 0, message: "bad revocation proof format")
        let digest = HashAlgorithm.SHA2_256.hash(record.revocationProof[0])
        assert(digest.length == subjRec.ac.length, message: "wrong digest size")

        var match = true
        for idx, val in digest {
            if subjRec.ac[idx] != val {
                match = false
                break
            }
        }
        assert(match, message: "invalid revocation proof")

        // update data asset counters

        let decrementCounter = fun (dataAssetID: String) {
            var counter = self.dataAssetCounters[dataAssetID]
            if counter != nil && counter! > 0 {
                self.dataAssetCounters[dataAssetID] = counter! - 1
            }
        }

        for id in subjRec.dataAssets {
            decrementCounter(id)
        }
        decrementCounter(subjRec.address)

        // update subject record state

        subjRec.setStatus(val: "revoked")
        self.records[record.subjectRecord] = subjRec

        record.setStatus(val: "published")
        record.setBlockNumber(val: blockNumber)
        self.records[record.id] = record

        emit RecordRevoked(id: subjRec.id, blockID: self.currentBlockNumber)
    }

    access(contract)
    fun submitAssetHead(record: Record, blockNumber: UInt64) {
        let prevHeadRecordID = self.assetHeads[record.headID]

        if record.subjectRecord != "" {
            assert(prevHeadRecordID != nil, message: "invalid reference to previous head")
            assert(prevHeadRecordID! == record.subjectRecord, message: "previous head mismatch")
            assert(record.revocationProof.length > 0, message: "bad revocation proof format")

            let subjRec = self.records[record.subjectRecord]
                ?? panic("previous asset head record not found: ".concat(record.subjectRecord))

            // check revocation proof

            let digest = HashAlgorithm.SHA2_256.hash(record.revocationProof[0])
            assert(digest.length == subjRec.ac.length, message: "wrong digest size")

            var match = true
            for idx, val in digest {
                if subjRec.ac[idx] != val {
                    match = false
                    break
                }
            }
            assert(match, message: "asset head revocation failed: invalid proof")

            // update subject record state
            subjRec.setStatus(val: "revoked")
            self.records[record.subjectRecord] = subjRec

            emit RecordRevoked(id: subjRec.id, blockID: self.currentBlockNumber)
        } else {
            assert(prevHeadRecordID == nil, message: "missing reference to previous head")
        }

        record.setStatus(val: "published")
        record.setBlockNumber(val: blockNumber)
        self.records[record.id] = record
        self.assetHeads[record.headID] = record.id
    }

    access(all)
    fun importBlock(number: UInt64, records: [Record]) {
        pre {
            self.currentBlockNumber == number-1 : "unexpected current block number"
        }
        self.rollBlock(blockHeight: getCurrentBlock().height)
        for rec in records {
            self.submitRecord(record: rec)
        }
    }

    access(all)
    fun getRecord(id: String): Record? {
        return self.records[id]
    }

    access(all)
    fun getDataAssetCounter(id: String): UInt64? {
        let counter = self.dataAssetCounters[id]
        if counter != nil {
            return counter!
        } else {
            return nil
        }
    }

    access(all)
    fun getAssetHead(headID: String): Record? {
        let rid = self.assetHeads[headID]
        if rid != nil {
            return self.records[rid!]
        } else {
            return nil
        }
    }

    access(all)
    fun getBlock(number: UInt64): MetaBlock? {
        return self.blocks[number]
    }

    access(all)
    fun getBlockRecords(blockNumber: UInt64): [String] {
        return []
    }

    access(all)
    fun getTopBlock(): UInt64 {
        return self.currentBlockNumber
    }

    access(all)
    fun getGenesisBlockHeight(): UInt64 {
        return self.genesisBlockHeight
    }

    access(all)
    fun getTopBlockHeight(): UInt64 {
        return self.currentBlockHeight
    }

    // initializer
    //
    init() {
        self.blocks = {}
        self.records = {}
        self.dataAssetCounters = {}
        self.assetHeads = {}
        self.prevBlockHash = []
        self.currentBlockNumber = 0
        self.currentRecordCount = 0
        self.currentBlockBody = []
        self.currentBlockHeight = getCurrentBlock().height
        self.genesisBlockHeight = self.currentBlockHeight

        emit MetaLockerInitialized()

        // create genesis block

        self.blocks[self.currentBlockNumber] = MetaBlock(number: self.currentBlockNumber, parentHash: "", status: 1)
        emit BlockOpened(chain: self.account.address, id: self.currentBlockNumber, parentHash: "")
        self.rollBlock(blockHeight: self.currentBlockHeight)
        //emit BlockConfirmed(chain: self.account.address, id: 0, parentHash: "", nonce: "", recordCount: 0, hash: "")
    }
}
