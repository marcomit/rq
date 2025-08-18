// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.

package request

type FtpCommand struct {
	name     string // Name of the command
	required int8   // Number of requires params (-1 for dynamic params)
}

var ftpCommands []FtpCommand = []FtpCommand{
	// Core FTP Commands (RFC 959)
	FtpCommand{"USER", 1}, // Send username for authentication
	FtpCommand{"PASS", 1}, // Send the password for authentication (after `USER`)
	FtpCommand{"ACCT", 1}, // Provide account info (rarely used)
	FtpCommand{"CWD", 1},  // Change working directory
	FtpCommand{"CDUP", 0}, // Go to parent directory
	FtpCommand{"SMNT", 1}, // Mount a different file system (obsolete).
	FtpCommand{"QUIT", 0}, // Terminate session.
	FtpCommand{"REIN", 0}, // Reinitialize connection (log out without closing TCP)

	// File Operations
	FtpCommand{"RETR", 1},  // Download (retrieve) a file
	FtpCommand{"STOR", 1},  // Upload (store) a file
	FtpCommand{"STOU", 0},  // Upload with a unique filename (server chooses)
	FtpCommand{"APPE", 1},  // Append to a file
	FtpCommand{"DELE", 1},  // Delete a file
	FtpCommand{"RNFR", 1},  // Specify file to rename (rename-from)
	FtpCommand{"RNTO", 1},  // New name for file (rename-to)
	FtpCommand{"MKD", 1},   // Make directory
	FtpCommand{"RMD", 1},   // Remove directory
	FtpCommand{"PWD", 1},   // Print working directory
	FtpCommand{"LIST", -1}, // List files and directories (detailed)
	FtpCommand{"NLST", -1}, // List filenames only

	// Transfer Parameters
	FtpCommand{"TYPE A", 0}, // Set transfer type to ASCII
	FtpCommand{"TYPR I", 0}, // Set transfer type to binary (image)
	FtpCommand{"STRU F", 1}, // Set file structore (F = File, mnost common)
	FtpCommand{"MODE S", 1}, // Set transfer mode (S = Stream, most common)

	// Data Connection Management
	FtpCommand{"PORT", 0}, // Active mode: client tells server where to connect for data
	FtpCommand{"PASV", 0}, // Passive mode: server tells client where to connect for data
	FtpCommand{"EPSV", 0}, // Extensive passive (modern, IPv6-friendly)
	FtpCommand{"EPRT", 0}, // Extended active mode

	// Information / Status
	FtpCommand{"SYST", 0},  // Ask server OS type
	FtpCommand{"STAT", -1}, // Get status of server or file
	FtpCommand{"HELP", -1}, // Get help from server
	FtpCommand{"NOOP", 0},  // Do nothing (no-operation)

	// Security (Extensions)
	FtpCommand{"AUTH TLS", 0}, // Start TLS encryption (FTPS)
	FtpCommand{"PBSZ 0", 0},   // Protection buffer size (TLS)
	FtpCommand{"PROT P", 0},   // Set protection to Private (TLS-encrypted data)

	// Common Non-Standard Extensions
	FtpCommand{"FEAT", 0},  // List server features (extensions supported)
	FtpCommand{"OPTS", -1}, // Set options for a feature
	FtpCommand{"SIZE", 1},  // Get size of a file
	FtpCommand{"MDTM", 1},  // Get last modified timestamp of a file
	FtpCommand{"MLSD", -1}, // Machine-readable listing of a directory contents
	FtpCommand{"MLST", 1},  // Machine-readable listing of a single file
}

func FtpTemplate() string {
	return `# The first line is the url of connection and the rest is the bytes to send
ftp.example.com:21
USER anonymous
PASS guest@example.com
PWD
CWD /pub
LIST
RETR example.txt
STOR upload.txt
QUIT
		`

}
