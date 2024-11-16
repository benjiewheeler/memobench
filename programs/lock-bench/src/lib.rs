use anchor_lang::prelude::*;

declare_id!("memos5eKsok4hXwxufkppaTBrfWRjFMhqxgomv5trJ6");

#[program]
pub mod lock_bench {
    use super::*;

    use {
        solana_program::{msg, program_error::ProgramError},
        std::str::from_utf8,
    };

    pub fn lock(ctx: Context<Lock>, input: Vec<u8>) -> Result<()> {
        let memo = from_utf8(&input).map_err(|err| {
            msg!("Invalid UTF-8, from byte {}", err.valid_up_to());
            ProgramError::InvalidInstructionData
        })?;

        let counter = &mut ctx.accounts.counter;
        counter.count = counter.count.checked_add(1).unwrap_or(u64::MAX);

        msg!("memobench(Lock): {} n={}", memo, counter.count);

        Ok(())
    }

    pub fn memo(_ctx: Context<Memo>, input: Vec<u8>) -> Result<()> {
        let memo = from_utf8(&input).map_err(|err| {
            msg!("Invalid UTF-8, from byte {}", err.valid_up_to());
            ProgramError::InvalidInstructionData
        })?;
        msg!("memobench(Memo): {}", memo);

        Ok(())
    }
}

#[derive(Accounts)]
pub struct Lock<'info> {
    #[account(mut)]
    pub payer: Signer<'info>,
    #[account(
        init_if_needed,
        payer = payer,
        space = 8 + 8,
        seeds = [b"counter"],
        bump,
     )]
    pub counter: Account<'info, Counter>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct Memo {}

#[account]
pub struct Counter {
    pub count: u64,
}

