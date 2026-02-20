import { PropsWithChildren } from "react";

interface Props extends PropsWithChildren {
  title: string;
  desc?: string;
}

export function SectionCard({ title, desc, children }: Props) {
  return (
    <section className="card">
      <div className="card-header">
        <h2>{title}</h2>
        {desc ? <p>{desc}</p> : null}
      </div>
      {children}
    </section>
  );
}


