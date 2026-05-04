export interface CommentReply {
  id: string;
  author: string;
  text: string;
  timestamp: string;
}

export interface CommentThread {
  id: string;
  quote: string;
  replies: CommentReply[];
  resolved: boolean;
}

class CommentStore {
  threads = $state<CommentThread[]>([]);
  activeCommentId = $state<string | null>(null);

  constructor() {
    // Basic mock initialization for design phase
    // In actual implementation, these will come from .kitap metadata
  }

  addThread(id: string, quote: string, initialText: string, author: string = 'Kullanıcı') {
    const newThread: CommentThread = {
      id,
      quote,
      resolved: false,
      replies: [
        {
          id: crypto.randomUUID(),
          author,
          text: initialText,
          timestamp: new Date().toISOString()
        }
      ]
    };
    this.threads.unshift(newThread);
    this.activeCommentId = id;
  }

  addReply(threadId: string, text: string, author: string = 'Kullanıcı') {
    const thread = this.threads.find(t => t.id === threadId);
    if (thread) {
      thread.replies.push({
        id: crypto.randomUUID(),
        author,
        text,
        timestamp: new Date().toISOString()
      });
    }
  }

  resolveThread(threadId: string) {
    const thread = this.threads.find(t => t.id === threadId);
    if (thread) thread.resolved = true;
  }
  
  removeThread(threadId: string) {
    this.threads = this.threads.filter(t => t.id !== threadId);
    if (this.activeCommentId === threadId) this.activeCommentId = null;
  }

  setActive(threadId: string | null) {
    this.activeCommentId = threadId;
  }
}

export const commentStore = new CommentStore();
