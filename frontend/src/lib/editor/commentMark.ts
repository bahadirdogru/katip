import { Mark, mergeAttributes } from '@tiptap/core';

export interface CommentOptions {
  HTMLAttributes: Record<string, any>;
}

declare module '@tiptap/core' {
  interface Commands<ReturnType> {
    comment: {
      /**
       * Set a comment mark
       */
      setComment: (commentId: string) => ReturnType;
      /**
       * Unset a comment mark
       */
      unsetComment: (commentId: string) => ReturnType;
    };
  }
}

export const CommentMark = Mark.create<CommentOptions>({
  name: 'comment',
  
  addOptions() {
    return {
      HTMLAttributes: {},
    };
  },

  inclusive: false,

  addAttributes() {
    return {
      commentId: {
        default: null,
        parseHTML: element => element.getAttribute('data-comment-id'),
        renderHTML: attributes => {
          if (!attributes.commentId) {
            return {};
          }
          return {
            'data-comment-id': attributes.commentId,
            class: 'comment-mark',
          };
        },
      },
    };
  },

  parseHTML() {
    return [
      {
        tag: 'span[data-comment-id]',
      },
    ];
  },

  renderHTML({ HTMLAttributes }) {
    return ['span', mergeAttributes(this.options.HTMLAttributes, HTMLAttributes), 0];
  },

  addCommands() {
    return {
      setComment:
        (commentId) =>
        ({ commands }) => {
          return commands.setMark(this.name, { commentId });
        },
      unsetComment:
        (commentId) =>
        ({ tr, dispatch }) => {
          if (!dispatch) return false;
          
          let hasMark = false;
          tr.doc.descendants((node, pos) => {
            if (node.isText) {
              const mark = node.marks.find(m => m.type.name === this.name && m.attrs.commentId === commentId);
              if (mark) {
                hasMark = true;
                tr.removeMark(pos, pos + node.nodeSize, mark.type);
              }
            }
          });
          
          return hasMark;
        },
    };
  },
});
