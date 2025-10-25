import { useMemo, useState } from 'react';

type ConversionResult = {
  base: string;
  target: string;
  amount: number;
  rate: number;
  converted: number;
  source: string;
};

const currencies = ['USD', 'IDR', 'SGD', 'JPY', 'KRW', 'VND'] as const;

type Currency = (typeof currencies)[number];

const defaultBase: Currency = 'USD';
const defaultTarget: Currency = 'IDR';

export default function App() {
  const [base, setBase] = useState<Currency>(defaultBase);
  const [target, setTarget] = useState<Currency>(defaultTarget);
  const [amount, setAmount] = useState<string>('1');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ConversionResult | null>(null);

  const isSwapDisabled = useMemo(() => base === target, [base, target]);

  const handleConvert = async () => {
    setLoading(true);
    setError(null);
    setResult(null);

    const numericAmount = Number(amount);
    if (Number.isNaN(numericAmount) || numericAmount <= 0) {
      setError('Please enter a positive number for the amount.');
      setLoading(false);
      return;
    }

    try {
      const params = new URLSearchParams({
        base,
        target,
        amount: numericAmount.toString()
      });
      const response = await fetch(`/api/convert?${params.toString()}`);
      if (!response.ok) {
        throw new Error('Conversion failed. Please try again.');
      }
      const payload: ConversionResult = await response.json();
      setResult(payload);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('Unknown error occurred.');
      }
    } finally {
      setLoading(false);
    }
  };

  const swapCurrencies = () => {
    setBase(target);
    setTarget(base);
  };

  return (
    <div className="app-container">
      <header>
        <h1>Currency Converter</h1>
        <p>Convert between supported currencies with live Yahoo Finance rates.</p>
      </header>

      <section className="card">
        <div className="field-group">
          <label htmlFor="amount">Amount</label>
          <input
            id="amount"
            type="number"
            inputMode="decimal"
            min="0"
            value={amount}
            onChange={(event) => setAmount(event.target.value)}
          />
        </div>

        <div className="field-row">
          <div className="field-group">
            <label htmlFor="base">From</label>
            <select id="base" value={base} onChange={(event) => setBase(event.target.value as Currency)}>
              {currencies.map((currency) => (
                <option key={currency} value={currency}>
                  {currency}
                </option>
              ))}
            </select>
          </div>

          <button className="swap" onClick={swapCurrencies} disabled={isSwapDisabled} title="Swap currencies">
            ⇄
          </button>

          <div className="field-group">
            <label htmlFor="target">To</label>
            <select id="target" value={target} onChange={(event) => setTarget(event.target.value as Currency)}>
              {currencies.map((currency) => (
                <option key={currency} value={currency}>
                  {currency}
                </option>
              ))}
            </select>
          </div>
        </div>

        <button className="convert" onClick={handleConvert} disabled={loading}>
          {loading ? 'Converting…' : 'Convert'}
        </button>

        {error && <p className="error">{error}</p>}
        {result && !error && (
          <div className="result" role="status">
            <p>
              {result.amount.toLocaleString(undefined, { maximumFractionDigits: 2 })} {result.base} =
            </p>
            <p className="converted-value">
              {result.converted.toLocaleString(undefined, { maximumFractionDigits: 2 })} {result.target}
            </p>
            <p className="rate">Rate: {result.rate.toFixed(4)} ({result.source})</p>
          </div>
        )}
      </section>

      <footer>
        <small>Backend proxy fetches data from Yahoo Finance.</small>
      </footer>
    </div>
  );
}
