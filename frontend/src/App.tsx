import { useState } from 'react';
import './App.css';
import { GenerateQRCode, SaveQRCode, SelectLogoFile } from '../wailsjs/go/main/App';

interface QRParams {
  url: string;
  size: number;
  fgColor: string;
  bgColor: string;
  logoPath: string;
  errorLevel: string;
  format: string;
  disableBorder: boolean;
}

const ERROR_LEVELS = [
  { value: 'H', label: 'H – Alto (30%)' },
  { value: 'Q', label: 'Q – Buono (25%)' },
  { value: 'M', label: 'M – Medio (15%)' },
  { value: 'L', label: 'L – Basso (7%)' },
];

function App() {
  const [params, setParams] = useState<QRParams>({
    url: '',
    size: 512,
    fgColor: '#000000',
    bgColor: '#ffffff',
    logoPath: '',
    errorLevel: 'H',
    format: 'png',
    disableBorder: false,
  });
  const [qrBase64, setQrBase64] = useState('');
  const [logoName, setLogoName] = useState('');
  const [error, setError] = useState('');
  const [savedPath, setSavedPath] = useState('');
  const [saving, setSaving] = useState(false);
  const [generating, setGenerating] = useState(false);

  function update<K extends keyof QRParams>(key: K, value: QRParams[K]) {
    setParams(prev => ({ ...prev, [key]: value }));
    setSavedPath('');
  }

  async function handleGenerate() {
    if (!params.url.trim()) {
      setError("L'URL è obbligatorio.");
      return;
    }
    setError('');
    setSavedPath('');
    setGenerating(true);
    try {
      const b64 = await GenerateQRCode(params);
      setQrBase64(b64);
    } catch (e: any) {
      setError(String(e));
    } finally {
      setGenerating(false);
    }
  }

  async function handleSelectLogo() {
    try {
      const path = await SelectLogoFile();
      if (path) {
        update('logoPath', path);
        // Handle both Unix (/) and Windows (\) path separators
        const parts = path.split(/[/\\]/);
        setLogoName(parts[parts.length - 1] || path);
      }
    } catch (e: any) {
      setError(String(e));
    }
  }

  function handleRemoveLogo() {
    setParams(prev => ({ ...prev, logoPath: '' }));
    setLogoName('');
    setSavedPath('');
  }

  async function handleSave() {
    setSaving(true);
    setSavedPath('');
    try {
      const path = await SaveQRCode(params);
      if (path) {
        const parts = path.split(/[/\\]/);
        setSavedPath(parts[parts.length - 1] || path);
      }
    } catch (e: any) {
      setError(String(e));
    } finally {
      setSaving(false);
    }
  }

  const logoLocked = params.logoPath !== '';

  return (
    <div className="layout">
      <aside className="panel">
        <h1 className="title">QR Code Generator</h1>

        <div className="field">
          <label className="label" htmlFor="url">URL <span className="required">*</span></label>
          <input
            id="url"
            className="input"
            type="url"
            placeholder="https://esempio.com"
            value={params.url}
            onChange={e => update('url', e.target.value)}
            onKeyDown={e => e.key === 'Enter' && handleGenerate()}
          />
        </div>

        <div className="field">
          <label className="label" htmlFor="size">
            Dimensione: <span className="value-badge">{params.size}px</span>
          </label>
          <div className="slider-row">
            <span className="slider-bound">128</span>
            <input
              id="size"
              className="slider"
              type="range"
              min={128}
              max={1024}
              step={64}
              value={params.size}
              onChange={e => update('size', Number(e.target.value))}
            />
            <span className="slider-bound">1024</span>
          </div>
        </div>

        <div className="field color-row">
          <div className="color-field">
            <label className="label" htmlFor="fg">Colore QR</label>
            <div className="color-wrapper">
              <input
                id="fg"
                type="color"
                className="color-input"
                value={params.fgColor}
                onChange={e => update('fgColor', e.target.value)}
              />
              <span className="color-hex">{params.fgColor}</span>
            </div>
          </div>
          <div className="color-field">
            <label className="label" htmlFor="bg">Sfondo</label>
            <div className="color-wrapper">
              <input
                id="bg"
                type="color"
                className="color-input"
                value={params.bgColor}
                onChange={e => update('bgColor', e.target.value)}
              />
              <span className="color-hex">{params.bgColor}</span>
            </div>
          </div>
        </div>

        <div className="field two-col">
          <div className="sub-field">
            <label className="label" htmlFor="errorLevel">
              Correzione errori
              {logoLocked && <span className="lock-badge" title="Fisso a H quando è presente un logo"> H</span>}
            </label>
            <select
              id="errorLevel"
              className="select"
              value={params.errorLevel}
              onChange={e => update('errorLevel', e.target.value)}
              disabled={logoLocked}
            >
              {ERROR_LEVELS.map(l => (
                <option key={l.value} value={l.value}>{l.label}</option>
              ))}
            </select>
          </div>
          <div className="sub-field">
            <label className="label" htmlFor="format">Formato</label>
            <select
              id="format"
              className="select"
              value={params.format}
              onChange={e => update('format', e.target.value)}
            >
              <option value="png">PNG</option>
              <option value="jpeg">JPEG</option>
            </select>
          </div>
        </div>

        <div className="field">
          <label className="label">Logo (opzionale)</label>
          {logoName ? (
            <div className="logo-selected">
              <span className="logo-name">{logoName}</span>
              <button className="btn-ghost" onClick={handleRemoveLogo}>Rimuovi</button>
            </div>
          ) : (
            <button className="btn-secondary" onClick={handleSelectLogo}>
              Seleziona immagine…
            </button>
          )}
          {logoLocked && (
            <p className="info-msg">Con il logo, la correzione errori viene forzata al livello H.</p>
          )}
        </div>

        <label className="checkbox-row">
          <input
            type="checkbox"
            className="checkbox"
            checked={params.disableBorder}
            onChange={e => update('disableBorder', e.target.checked)}
          />
          <span className="checkbox-label">Rimuovi margine (quiet zone)</span>
        </label>

        {error && <p className="error-msg">{error}</p>}

        <button
          className="btn-primary"
          onClick={handleGenerate}
          disabled={generating}
        >
          {generating ? 'Generazione…' : 'Genera QR Code'}
        </button>
      </aside>

      <main className="preview">
        {qrBase64 ? (
          <>
            <img
              className="qr-image"
              src={`data:image/png;base64,${qrBase64}`}
              alt="QR Code generato"
            />
            <div className="save-row">
              <button
                className="btn-save"
                onClick={handleSave}
                disabled={saving}
              >
                {saving ? 'Salvataggio…' : `Salva come ${params.format.toUpperCase()}`}
              </button>
              {savedPath && (
                <span className="saved-badge">Salvato: {savedPath}</span>
              )}
            </div>
          </>
        ) : (
          <div className="placeholder">
            <div className="placeholder-icon">⬛</div>
            <p>Il QR Code apparirà qui</p>
          </div>
        )}
      </main>
    </div>
  );
}

export default App;
