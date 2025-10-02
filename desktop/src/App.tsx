import { useState, useEffect } from "react";
import { invoke } from "@tauri-apps/api/core";
import { open } from "@tauri-apps/plugin-dialog";
import "./App.css";

const API_BASE = "http://localhost:8765";

interface JobStatus {
  job_id: string;
  status: "pending" | "running" | "completed" | "failed";
  message?: string;
  error?: string;
  output_path?: string;
  stats?: {
    total_messages: number;
    text_messages: number;
    total_contacts: number;
    attachment_count: number;
  };
}

function App() {
  const [dbPath, setDbPath] = useState("");
  const [defaultFound, setDefaultFound] = useState(false);
  const [includeAttachments, setIncludeAttachments] = useState(false);
  const [attachmentsWarning, setAttachmentsWarning] = useState(false);
  const [title, setTitle] = useState("Our Messages");
  const [outputFormat, setOutputFormat] = useState("tex");
  const [jobId, setJobId] = useState("");
  const [jobStatus, setJobStatus] = useState<JobStatus | null>(null);
  const [isGenerating, setIsGenerating] = useState(false);
  const [healthStatus, setHealthStatus] = useState<string>("checking...");

  // Check for default Messages path and API health on mount
  useEffect(() => {
    checkDefaultPath();
    checkHealth();
  }, []);

  async function checkDefaultPath() {
    try {
      const path = await invoke<string | null>("check_default_messages_path");
      if (path) {
        setDbPath(path);
        setDefaultFound(true);
      }
    } catch (error) {
      console.error("Failed to check default path:", error);
    }
  }

  async function selectDirectory() {
    try {
      const selected = await open({
        directory: true,
        multiple: false,
        title: "Select Messages Directory",
      });

      if (selected && typeof selected === "string") {
        // Append chat.db to the selected directory
        const dbPath = selected + "/chat.db";
        setDbPath(dbPath);
        setDefaultFound(true); // After manual selection, show "Change"
      }
    } catch (error) {
      console.error("Failed to open directory picker:", error);
    }
  }

  async function checkAttachmentsDirectory() {
    if (!includeAttachments || !dbPath) {
      setAttachmentsWarning(false);
      return;
    }

    try {
      // Get the directory containing chat.db
      const dbDir = dbPath.substring(0, dbPath.lastIndexOf("/"));
      const attachmentsPath = dbDir + "/Attachments";

      const exists = await invoke<boolean>("check_directory_exists", { path: attachmentsPath });
      setAttachmentsWarning(!exists);
    } catch (error) {
      console.error("Failed to check attachments directory:", error);
      setAttachmentsWarning(true);
    }
  }

  // Check for attachments directory when checkbox is toggled or dbPath changes
  useEffect(() => {
    checkAttachmentsDirectory();
  }, [includeAttachments, dbPath]);

  async function checkHealth() {
    try {
      const response = await fetch(`${API_BASE}/api/health`);
      if (response.ok) {
        setHealthStatus("connected");
      } else {
        setHealthStatus("error");
      }
    } catch (error) {
      setHealthStatus("disconnected");
      console.error("Health check failed:", error);
    }
  }

  async function generateBook() {
    if (!dbPath) {
      alert("Please enter the path to your iMessages database");
      return;
    }

    setIsGenerating(true);
    setJobStatus(null);

    try {
      // Get the directory containing chat.db
      const dbDir = dbPath.substring(0, dbPath.lastIndexOf("/"));
      const attachmentsPath = includeAttachments ? dbDir + "/Attachments" : "";

      const response = await fetch(`${API_BASE}/api/generate`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          database_path: dbPath,
          attachments_path: attachmentsPath || "Attachments",
          output_path: `book.${outputFormat}`,
          title: title,
          include_images: includeAttachments,
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to start generation");
      }

      const data = await response.json();
      setJobId(data.job_id);

      // Start polling for status
      pollJobStatus(data.job_id);
    } catch (error) {
      console.error("Failed to generate book:", error);
      alert("Failed to start book generation. Make sure the backend is running.");
      setIsGenerating(false);
    }
  }

  async function pollJobStatus(id: string) {
    const interval = setInterval(async () => {
      try {
        const response = await fetch(`${API_BASE}/api/jobs/${id}`);
        if (!response.ok) {
          throw new Error("Failed to fetch job status");
        }

        const status: JobStatus = await response.json();
        setJobStatus(status);

        if (status.status === "completed" || status.status === "failed") {
          clearInterval(interval);
          setIsGenerating(false);
        }
      } catch (error) {
        console.error("Failed to fetch job status:", error);
        clearInterval(interval);
        setIsGenerating(false);
      }
    }, 1000);
  }

  return (
    <main className="container">
      <h1>📖 Threadbound</h1>
      <p className="subtitle">Generate beautiful books from your iMessages</p>

      <div className="status-indicator">
        <span className={`status-dot ${healthStatus}`}></span>
        Backend: {healthStatus}
      </div>

      <div className="form-section">
        <h2>Settings</h2>

        <div className="form-group">
          <label htmlFor="db-path">Messages Location *</label>
          <div className="input-with-button">
            <input
              id="db-path"
              type="text"
              value={dbPath}
              onChange={(e) => setDbPath(e.target.value)}
              placeholder="~/Library/Messages/chat.db"
            />
            <button
              type="button"
              onClick={selectDirectory}
              className="select-btn"
            >
              {defaultFound ? "Change" : "Select"}
            </button>
          </div>
        </div>

        <div className="form-group">
          <label>
            <input
              type="checkbox"
              checked={includeAttachments}
              onChange={(e) => setIncludeAttachments(e.target.checked)}
            />
            {" "}Include Attachments
          </label>
          {attachmentsWarning && includeAttachments && (
            <div className="warning-message">
              ⚠️ No Attachments directory found in the same folder as chat.db
            </div>
          )}
        </div>

        <div className="form-group">
          <label htmlFor="title">Book Title</label>
          <input
            id="title"
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Our Messages"
          />
        </div>

        <div className="form-group">
          <label htmlFor="format">Output Format</label>
          <select
            id="format"
            value={outputFormat}
            onChange={(e) => setOutputFormat(e.target.value)}
          >
            <option value="tex">TeX (for PDF generation)</option>
            <option value="html">HTML</option>
            <option value="txt">Text</option>
          </select>
        </div>

        <button
          onClick={generateBook}
          disabled={isGenerating || !dbPath}
          className="generate-btn"
        >
          {isGenerating ? "Generating..." : "Generate Book"}
        </button>
      </div>

      {jobStatus && (
        <div className="job-status">
          <h2>Generation Status</h2>
          <div className={`status-card ${jobStatus.status}`}>
            <div className="status-header">
              <span className="status-label">{jobStatus.status}</span>
              <span className="job-id">Job: {jobId.slice(0, 8)}...</span>
            </div>

            {jobStatus.message && <p>{jobStatus.message}</p>}

            {jobStatus.error && (
              <div className="error-message">
                <strong>Error:</strong> {jobStatus.error}
              </div>
            )}

            {jobStatus.stats && (
              <div className="stats">
                <h3>📊 Statistics</h3>
                <ul>
                  <li>Messages: {jobStatus.stats.total_messages.toLocaleString()}</li>
                  <li>Text Messages: {jobStatus.stats.text_messages.toLocaleString()}</li>
                  <li>Contacts: {jobStatus.stats.total_contacts}</li>
                  <li>Attachments: {jobStatus.stats.attachment_count}</li>
                </ul>
              </div>
            )}

            {jobStatus.output_path && jobStatus.status === "completed" && (
              <div className="success-message">
                <strong>✅ Book generated successfully!</strong>
                <p>Output: {jobStatus.output_path}</p>
              </div>
            )}
          </div>
        </div>
      )}
    </main>
  );
}

export default App;
