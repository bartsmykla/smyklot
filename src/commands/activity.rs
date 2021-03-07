use serenity::{
    prelude::*,
    model::{
        channel::{Message},
        gateway::{Activity},
    },
    framework::standard::{
        Args, CommandResult,
        macros::*,
    },
};

#[command]
#[bucket = "activity"]
#[aliases("set")]
async fn play(ctx: &Context, _msg: &Message, args: Args) -> CommandResult {
    let name = args.message();

    ctx.set_activity(Activity::playing(&name)).await;

    Ok(())
}
