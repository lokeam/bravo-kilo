export interface DeltaOp {
  insert: string;
  attributes?: {
    bold?: boolean;
    italic?: boolean;
    underline?: boolean;
    header?: number;
    list?: 'ordered' | 'bullet';
    align?: 'center' | 'right' | 'justify';
    link?: string;
    blockquote?: boolean;
    code?: boolean;
    script?: 'sub' | 'super';
    color?: string;
    background?: string;
  };
}

export interface Block {
  type: string;
  level?: number;
  content: string;
}