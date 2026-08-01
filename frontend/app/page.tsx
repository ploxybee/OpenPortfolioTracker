"use client";

import { useEffect, useState } from "react";

type Holding = {
  symbol: string;
  name: string;
  account: string;
  assetClass: string;
  country: string;
  currency: string;
  marketValue: number;
  unrealizedPnlPercent: number;
  weight: number;
};
type Allocation = {
  label: string;
  value: number;
  percentage: number;
  targetPercentage?: number;
  driftPercentage?: number;
};
type Suggestion = { assetClass: string; action: string; amount: number };
type Snapshot = {
  broker: string;
  mode: string;
  updatedAt: string;
  totalValue: number;
  currency: string;
  holdings: Holding[];
  assetAllocation: Allocation[];
  targets: Record<string, number>;
  rebalanceTolerance: number;
  monthlyContribution: number;
  contributionPlan?: Suggestion[] | null;
};

const api = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
const money = (value: number, currency: string) =>
  new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(value);
const pct = (value: number) => `${value.toFixed(1)}%`;

export default function Home() {
  const [data, setData] = useState<Snapshot>();
  const [error, setError] = useState<string>();
  const load = async () => {
    setError(undefined);
    try {
      const response = await fetch(`${api}/api/v1/portfolio`);
      if (!response.ok) throw new Error("Portfolio unavailable");
      setData(await response.json());
    } catch {
      setError(
        "Couldn’t reach the portfolio service. Start the API, then refresh.",
      );
    }
  };
  useEffect(() => {
    void load();
  }, []);
  if (error)
    return (
      <main className="shell">
        <p className="eyebrow">OPEN PORTFOLIO TRACKER</p>
        <h1>Your dashboard is waiting.</h1>
        <p>{error}</p>
        <button onClick={() => void load()}>Try again</button>
      </main>
    );
  if (!data)
    return (
      <main className="shell">
        <p className="eyebrow">OPEN PORTFOLIO TRACKER</p>
        <h1>Loading your portfolio…</h1>
      </main>
    );
  const contributionPlan = data.contributionPlan ?? [];
  return (
    <main className="shell">
      <header>
        <div>
          <p className="eyebrow">OPEN PORTFOLIO TRACKER</p>
          <h1>Invest with perspective.</h1>
          <p className="subtle">
            A low-effort view of your {data.broker} portfolio and plan.
          </p>
        </div>
        <button onClick={() => void load()}>Refresh</button>
      </header>
      {data.mode === "demo" && (
        <aside>
          <strong>Demo portfolio</strong>
          <span>
            Add broker credentials to <code>backend/.env</code> to see live
            holdings.
          </span>
        </aside>
      )}
      <section className="hero">
        <p>Total portfolio value</p>
        <h2>{money(data.totalValue, data.currency)}</h2>
        <span>
          Updated{" "}
          {new Intl.DateTimeFormat("en", {
            dateStyle: "medium",
            timeStyle: "short",
          }).format(new Date(data.updatedAt))}
        </span>
      </section>
      <section className="grid">
        <AllocationCard
          items={data.assetAllocation}
          tolerance={data.rebalanceTolerance}
        />
        <section className="allocation">
          <p className="eyebrow">LOW-EFFORT ACTION</p>
          <h2>Next contribution</h2>
          <p>
            {money(data.monthlyContribution, data.currency)} is allocated to the
            largest target gaps first; no sales are proposed.
          </p>
          {contributionPlan.length ? (
            contributionPlan.map((item) => (
              <div className="allocation-row" key={item.assetClass}>
                <div>
                  <span>
                    {item.action} {item.assetClass}
                  </span>
                  <b>{money(item.amount, data.currency)}</b>
                </div>
              </div>
            ))
          ) : (
            <p className="subtle">No contribution plan configured.</p>
          )}
          <small>
            Review taxes, fees, and whole-share constraints before trading.
          </small>
        </section>
      </section>
      <section className="holdings">
        <div className="section-title">
          <div>
            <p className="eyebrow">CONSOLIDATED HOLDINGS</p>
            <h2>What you own</h2>
          </div>
          <span>{data.holdings.length} positions</span>
        </div>
        <div className="table">
          <div className="row heading">
            <span>Holding</span>
            <span>Class</span>
            <span>Weight</span>
            <span>Value</span>
          </div>
          {data.holdings.map((h) => (
            <div className="row" key={`${h.account}-${h.symbol}`}>
              <span>
                <b>{h.symbol}</b>
                <small>
                  {h.name} · {h.account}
                </small>
              </span>
              <span>{h.assetClass}</span>
              <span>{h.weight.toFixed(1)}%</span>
              <span>{money(h.marketValue, h.currency)}</span>
            </div>
          ))}
        </div>
      </section>
    </main>
  );
}
function AllocationCard({
  items,
  tolerance,
}: {
  items: Allocation[];
  tolerance: number;
}) {
  return (
    <section className="allocation">
      <p className="eyebrow">TARGET ALLOCATION</p>
      <h2>Actual versus plan</h2>
      {items.map((item) => {
        const drift = item.driftPercentage ?? 0;
        const status = Math.abs(drift) >= tolerance ? "attention" : "";
        return (
          <div className="allocation-row" key={item.label}>
            <div>
              <span>{item.label}</span>
              <b>{pct(item.percentage)}</b>
            </div>
            <div className="bar">
              <i style={{ width: `${item.percentage}%` }} />
              <em style={{ left: `${item.targetPercentage ?? 0}%` }} />
            </div>
            <small className={status}>
              Target {pct(item.targetPercentage ?? 0)} ·{" "}
              {drift >= 0 ? "over" : "under"} by {pct(Math.abs(drift))}
            </small>
          </div>
        );
      })}
    </section>
  );
}
