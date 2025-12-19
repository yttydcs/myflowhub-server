package file

import "encoding/json"

// 子协议：file（节点间文件传输）。
const SubProtoFile uint8 = 5

// payload[0]：帧类型。
const (
	kindCtrl byte = 0x01
	kindData byte = 0x02
	kindAck  byte = 0x03
)

const (
	actionRead      = "read"
	actionWrite     = "write"
	actionReadResp  = "read_resp"
	actionWriteResp = "write_resp"
)

const (
	opPull  = "pull"
	opOffer = "offer"
	opList  = "list"
	opReadText = "read_text"
)

type message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data"`
}

type readReq struct {
	Op         string `json:"op"`
	Target     uint32 `json:"target,omitempty"`
	Dir        string `json:"dir,omitempty"`
	Name       string `json:"name,omitempty"`
	Overwrite  *bool  `json:"overwrite,omitempty"`
	ResumeFrom uint64 `json:"resume_from,omitempty"`
	WantHash   *bool  `json:"want_hash,omitempty"`
	Recursive  bool   `json:"recursive,omitempty"`
	MaxBytes   uint32 `json:"max_bytes,omitempty"`
}

type readResp struct {
	Code      int      `json:"code"`
	Msg       string   `json:"msg,omitempty"`
	Op        string   `json:"op,omitempty"`
	SessionID string   `json:"session_id,omitempty"`
	Provider  uint32   `json:"provider,omitempty"`
	Consumer  uint32   `json:"consumer,omitempty"`
	Dir       string   `json:"dir,omitempty"`
	Name      string   `json:"name,omitempty"`
	Size      uint64   `json:"size,omitempty"`
	Sha256    string   `json:"sha256,omitempty"`
	StartFrom uint64   `json:"start_from,omitempty"`
	Chunk     uint32   `json:"chunk_bytes,omitempty"`
	Dirs      []string `json:"dirs,omitempty"`
	Files     []string `json:"files,omitempty"`
	Text      string   `json:"text,omitempty"`
	Truncated bool     `json:"truncated,omitempty"`
}

type writeReq struct {
	Op        string `json:"op"`
	Target    uint32 `json:"target"`
	SessionID string `json:"session_id"`
	Dir       string `json:"dir,omitempty"`
	Name      string `json:"name"`
	Size      uint64 `json:"size"`
	Sha256    string `json:"sha256,omitempty"`
	Overwrite *bool  `json:"overwrite,omitempty"`
}

type writeResp struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg,omitempty"`
	Op         string `json:"op,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
	Provider   uint32 `json:"provider,omitempty"`
	Consumer   uint32 `json:"consumer,omitempty"`
	Dir        string `json:"dir,omitempty"`
	Name       string `json:"name,omitempty"`
	Size       uint64 `json:"size,omitempty"`
	Sha256     string `json:"sha256,omitempty"`
	Accept     bool   `json:"accept,omitempty"`
	ResumeFrom uint64 `json:"resume_from,omitempty"`
}
