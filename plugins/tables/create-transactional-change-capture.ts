import { createTransactionChangeState } from "./transaction-change-state.js";
import { TransactionChangeWriter } from "./transaction-change-writer.js";
import { ScopedTransactionChangeWriter } from "./scoped-transaction-change-writer.js";
import { TransactionChangeReader } from "./transaction-change-reader.js";
import { ScopedTransactionChangeReader } from "./scoped-transaction-change-reader.js";
export function createTransactionalChangeCapture<T>() { const state = createTransactionChangeState<T>(); const writer = new TransactionChangeWriter(state); const scopedWriter = new ScopedTransactionChangeWriter(state); const reader = new TransactionChangeReader(state); const scopedReader = new ScopedTransactionChangeReader(state); return { begin: writer.begin.bind(writer), beginScoped: scopedWriter.beginScoped.bind(scopedWriter), list: reader.list.bind(reader), listScoped: scopedReader.listScoped.bind(scopedReader) }; }
