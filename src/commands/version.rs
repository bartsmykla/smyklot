use std::env;

use serenity::{
    prelude::*,
    model::{
        channel::{Message},
    },
    framework::standard::{
        CommandResult,
        macros::*,
    },
};

#[command]
async fn version(ctx: &Context, msg: &Message) -> CommandResult {
    let version = env::var("SMYKLOT_VERSION");

    let message = match version {
        Ok(v) if v != "{{version}}" => v,
        _ => String::from("¯\\_(ツ)_/¯")
    };

    msg.reply(ctx, message).await?;

    Ok(())
}
