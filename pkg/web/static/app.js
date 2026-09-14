let allChats = [];
let activeChat = null;

document.addEventListener('DOMContentLoaded', () => {
  loadAppInfo();
  loadChats();

  const searchInput = document.getElementById('chatSearchInput');
  searchInput.addEventListener('input', (e) => {
    filterChats(e.target.value.trim());
  });

  document.getElementById('btnBackMobile').addEventListener('click', () => {
    document.querySelector('.app-layout').classList.remove('chat-open');
  });

  document.getElementById('btnExportHtml').addEventListener('click', () => exportActiveChat('html'));
  document.getElementById('btnExportJson').addEventListener('click', () => exportActiveChat('json'));
  document.getElementById('btnExportTxt').addEventListener('click', () => exportActiveChat('txt'));
});

async function loadAppInfo() {
  try {
    const res = await fetch('/api/info');
    const info = await res.json();
    const badge = document.getElementById('dbStatusBadge');
    if (info.loaded) {
      badge.textContent = info.isModern ? 'Banco Moderno (2025) - ONLYCODE.COM.BR' : 'Banco Legado - ONLYCODE.COM.BR';
    } else {
      badge.textContent = 'Sem Banco';
      badge.style.color = '#f15c6d';
    }
  } catch (err) {
    console.error('Failed to load info', err);
  }
}

async function loadChats() {
  const container = document.getElementById('chatsList');
  try {
    const res = await fetch('/api/chats');
    if (!res.ok) {
      container.innerHTML = '<div class="loading-spinner">Erro ao carregar conversas.</div>';
      return;
    }
    allChats = await res.json();
    renderChats(allChats);
  } catch (err) {
    container.innerHTML = '<div class="loading-spinner">Erro de rede ao carregar conversas.</div>';
    console.error(err);
  }
}

function renderChats(chats) {
  const container = document.getElementById('chatsList');
  if (!chats || chats.length === 0) {
    container.innerHTML = '<div class="loading-spinner">Nenhuma conversa encontrada.</div>';
    return;
  }

  container.innerHTML = '';
  chats.forEach(chat => {
    const item = document.createElement('div');
    item.className = 'chat-item';
    if (activeChat && activeChat.jid === chat.jid) {
      item.classList.add('active');
    }

    const name = chat.displayName || chat.subject || chat.jid;
    const initials = getInitials(name);
    const dateStr = chat.lastMessageTimestamp ? formatDate(chat.lastMessageTimestamp) : '';

    item.innerHTML = `
      <div class="avatar">${escapeHtml(initials)}</div>
      <div class="chat-item-content">
        <div class="chat-item-header">
          <span class="chat-item-name">${escapeHtml(name)}</span>
          <span class="chat-item-time">${escapeHtml(dateStr)}</span>
        </div>
        <div class="chat-item-footer">
          <span class="chat-item-snippet">${escapeHtml(chat.jid)}</span>
          ${chat.totalCount > 0 ? `<span class="chat-item-badge">${chat.totalCount}</span>` : ''}
        </div>
      </div>
    `;

    item.addEventListener('click', () => selectChat(chat));
    container.appendChild(item);
  });
}

function filterChats(query) {
  if (!query) {
    renderChats(allChats);
    return;
  }
  const q = query.toLowerCase();
  const filtered = allChats.filter(c => {
    const name = (c.displayName || c.subject || c.jid).toLowerCase();
    const jid = c.jid.toLowerCase();
    return name.includes(q) || jid.includes(q);
  });
  renderChats(filtered);
}

async function selectChat(chat) {
  activeChat = chat;
  document.querySelector('.app-layout').classList.add('chat-open');

  // Highlight active
  document.querySelectorAll('.chat-item').forEach(el => el.classList.remove('active'));
  // Update header
  const name = chat.displayName || chat.subject || chat.jid;
  document.getElementById('activeChatTitle').textContent = name;
  document.getElementById('activeChatSubtitle').textContent = `${chat.jid} • ${chat.totalCount} mensagens`;
  document.getElementById('activeChatAvatar').textContent = getInitials(name);

  document.getElementById('emptyState').style.display = 'none';
  document.getElementById('activeChatWrapper').style.display = 'flex';

  const content = document.getElementById('messagesContent');
  content.innerHTML = '<div class="loading-spinner">Carregando mensagens...</div>';

  try {
    const res = await fetch(`/api/messages?jid=${encodeURIComponent(chat.jid)}`);
    if (!res.ok) {
      content.innerHTML = '<div class="loading-spinner">Erro ao obter mensagens.</div>';
      return;
    }
    const messages = await res.json();
    renderMessages(messages);
  } catch (err) {
    content.innerHTML = '<div class="loading-spinner">Erro de rede.</div>';
    console.error(err);
  }
}

function renderMessages(messages) {
  const content = document.getElementById('messagesContent');
  content.innerHTML = '';

  if (!messages || messages.length === 0) {
    content.innerHTML = '<div class="loading-spinner">Sem mensagens nesta conversa.</div>';
    return;
  }

  let lastDate = '';
  messages.forEach(msg => {
    if (msg.dateFormatted && msg.dateFormatted !== lastDate) {
      lastDate = msg.dateFormatted;
      const divider = document.createElement('div');
      divider.className = 'day-divider';
      divider.innerHTML = `<span class="day-chip">${escapeHtml(lastDate)}</span>`;
      content.appendChild(divider);
    }

    const row = document.createElement('div');
    row.className = `message-row ${msg.fromMe ? 'outgoing' : 'incoming'}`;

    const bubble = document.createElement('div');
    bubble.className = 'bubble';

    // Sender for group incoming
    if (!msg.fromMe && (msg.remoteResourceDisplayName || msg.remoteResource)) {
      const sender = msg.remoteResourceDisplayName || msg.remoteResource;
      const sDiv = document.createElement('div');
      sDiv.className = 'sender-name';
      sDiv.textContent = sender;
      bubble.appendChild(sDiv);
    }

    // Quote
    if (msg.quotedMessageId) {
      const quoteBox = document.createElement('div');
      quoteBox.className = 'quote-box';
      const author = msg.quotedMessage ? (msg.quotedMessage.fromMe ? 'Você' : (msg.quotedMessage.remoteResourceDisplayName || msg.quotedMessage.remoteResource || 'Remetente')) : 'Citação';
      const quoteSnippet = msg.quotedMessage ? (msg.quotedMessage.data || '[Mídia]') : msg.quotedMessageId;
      quoteBox.innerHTML = `
        <div class="quote-author">${escapeHtml(author)}</div>
        <div class="quote-snippet">${escapeHtml(quoteSnippet)}</div>
      `;
      bubble.appendChild(quoteBox);
    }

    // Media Thumbnail
    if (msg.thumbnailUrl) {
      const img = document.createElement('img');
      img.className = 'media-thumbnail';
      img.src = msg.thumbnailUrl;
      img.alt = 'thumbnail';
      bubble.appendChild(img);
    }

    // Body based on media type
    appendMediaContent(bubble, msg);

    // Meta (timestamp)
    const meta = document.createElement('span');
    meta.className = 'bubble-meta';
    meta.innerHTML = `
      ${escapeHtml(msg.timeFormatted || '')}
      ${msg.fromMe ? `<svg viewBox="0 0 16 15" width="16" height="15" fill="currentColor"><path d="M15.01 3.316l-.478-.372a.365.365 0 0 0-.51.063L8.666 9.879a.32.32 0 0 1-.484.033l-.358-.325a.319.319 0 0 0-.484.032l-.378.483a.418.418 0 0 0 .036.541l1.32 1.266c.143.14.361.125.484-.033l6.272-8.048a.366.366 0 0 0-.064-.512zm-4.1 0l-.478-.372a.365.365 0 0 0-.51.063L4.566 9.879a.32.32 0 0 1-.484.033L1.891 7.769a.366.366 0 0 0-.515.006l-.423.433a.364.364 0 0 0 .006.514l3.258 3.185c.143.14.361.125.484-.033l6.272-8.048a.365.365 0 0 0-.063-.51z"/></svg>` : ''}
    `;
    bubble.appendChild(meta);

    row.appendChild(bubble);
    content.appendChild(row);
  });

  // Scroll to bottom
  const viewport = document.getElementById('messagesViewport');
  viewport.scrollTop = viewport.scrollHeight;
}

function appendMediaContent(bubble, msg) {
  const t = msg.mediaWhatsappType;
  if (t === 1 || t === 13) { // Image or Gif
    if (msg.mediaCaption) {
      const cap = document.createElement('div');
      cap.style.marginTop = '4px';
      cap.textContent = msg.mediaCaption;
      bubble.appendChild(cap);
    }
  } else if (t === 2) { // Audio
    const badge = document.createElement('div');
    badge.className = 'media-badge';
    badge.textContent = `🎵 Áudio (${msg.mediaDuration || 0}s)`;
    bubble.appendChild(badge);
  } else if (t === 3) { // Video
    const badge = document.createElement('div');
    badge.className = 'media-badge';
    badge.textContent = `🎬 Vídeo ${msg.mediaName || ''}`;
    bubble.appendChild(badge);
    if (msg.mediaCaption) {
      const cap = document.createElement('div');
      cap.textContent = msg.mediaCaption;
      bubble.appendChild(cap);
    }
  } else if (t === 4) { // Contact
    const badge = document.createElement('div');
    badge.className = 'media-badge';
    badge.textContent = '👤 Contato';
    bubble.appendChild(badge);
  } else if (t === 5 || t === 16) { // Location
    const badge = document.createElement('div');
    badge.className = 'media-badge';
    badge.innerHTML = `📍 <a href="https://www.google.com/maps?q=${msg.latitude},${msg.longitude}" target="_blank" style="color:var(--text-accent);text-decoration:underline;">Localização (${msg.latitude.toFixed(4)}, ${msg.longitude.toFixed(4)})</a>`;
    bubble.appendChild(badge);
  } else if (t === 9) { // File
    const badge = document.createElement('div');
    badge.className = 'media-badge';
    badge.textContent = `📄 Documento: ${msg.mediaName || 'Arquivo'}`;
    bubble.appendChild(badge);
  } else {
    if (msg.isLink) {
      if (msg.mediaCaption) {
        const cap = document.createElement('div');
        cap.style.fontWeight = '600';
        cap.textContent = msg.mediaCaption;
        bubble.appendChild(cap);
      }
      const a = document.createElement('a');
      a.href = msg.data;
      a.target = '_blank';
      a.style.color = 'var(--text-accent)';
      a.textContent = msg.data;
      bubble.appendChild(a);
    } else {
      const textNode = document.createElement('div');
      textNode.style.whiteSpace = 'pre-wrap';
      textNode.textContent = msg.data;
      bubble.appendChild(textNode);
    }
  }
}

function exportActiveChat(format) {
  if (!activeChat) return;
  const url = `/api/export?jid=${encodeURIComponent(activeChat.jid)}&format=${format}`;
  window.open(url, '_blank');
}

function formatDate(ts) {
  const d = new Date(ts);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }
  return d.toLocaleDateString([], { day: '2-digit', month: '2-digit', year: '2-digit' });
}

function getInitials(name) {
  if (!name) return 'WA';
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) return parts[0].substring(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

function escapeHtml(str) {
  if (!str) return '';
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}
