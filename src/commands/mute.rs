use std::sync::Arc;
use serenity::{
    prelude::*,
    model::{
        channel::{Message},
        id::UserId,
        guild::{Member, Guild},
    },
    framework::standard::{
        Args, CommandResult,
        macros::*,
    },
};

use crate::Config;

#[command]
#[delimiters(" ")]
#[min_args(1)]
#[max_args(1)]
async fn mute(ctx: &Context, msg: &Message, mut args: Args) -> CommandResult {
    let config_lock = get_config_lock(ctx).await;
    let config = config_lock.read().await;
    let mute_role_id = config.mute_role_id.unwrap();
    let guild = get_guild(&ctx, msg).await?;
    let mut member = get_member(&ctx, &mut args, guild).await?;
    
    let message = match member.add_role(&ctx.http, mute_role_id).await {
        Ok(_) => format!("{} was muted", member),
        Err(err) => format!("Couldn't mute {}: {}", member, err),
    };
    
    msg.reply(&ctx.http, message).await?;
    
    Ok(())
}

#[command]
#[delimiters(" ")]
#[min_args(1)]
#[max_args(1)]
async fn unmute(ctx: &Context, msg: &Message, mut args: Args) -> CommandResult {
    let config_lock = get_config_lock(ctx).await;
    let config = config_lock.read().await;
    let guild = get_guild(&ctx, msg).await?;
    let mute_role_id = config.mute_role_id.unwrap();
    let mut member = get_member(&ctx, &mut args, guild).await?;
    
    let message = match member.remove_role(&ctx.http, mute_role_id).await {
        Ok(_) => format!("{} was unmuted", member),
        Err(err) => format!("Couldn't unmute {}: {}", member, err),
    };
    
    msg.reply(&ctx.http, message).await?;

    Ok(())
}

#[command]
async fn muted(ctx: &Context, msg: &Message) -> CommandResult {
    let config_lock = get_config_lock(ctx).await;
    let config = config_lock.read().await;
    let guild = get_guild(&ctx, msg).await?;
    let mute_role_id = config.mute_role_id.unwrap();
    let members = guild.members
        .iter()
        .filter(|(_, member)| member.roles.contains(&mute_role_id))
        .map(|(_, member)| member.to_string())
        .collect::<Vec<String>>();

    let message = if members.len() > 0 {
        format!("Currently muted members: {}", members.join(", "))
    } else {
        format!("No members are currently muted")
    };

    msg.reply(&ctx.http, message).await?;

    Ok(())
}

async fn get_config_lock(ctx: &Context) -> Arc<RwLock<Config>> {
    ctx.data.read().await
        .get::<Config>()
        .expect("Missing Config in Context")
        .clone()
}

async fn get_guild(ctx: &&Context, msg: &Message) -> Result<Guild, String> {
    let guild = msg.channel_id
        .to_channel(&ctx.http).await
        .map_err(|err| err.to_string())?
        .guild()
        .ok_or("Command works only in channels")?
        .guild(&ctx.cache).await
        .ok_or("nope")?;
    
    Ok(guild)
}

async fn get_member(ctx: &&Context, args: &mut Args, guild: Guild) -> Result<Member, String> {
    match args.single::<UserId>() {
        Ok(user_id) => guild.member(&ctx.http, user_id).await.map_err(|err| err.to_string()),
        Err(_) => {
            let user_name = args.current().unwrap();

            guild
                .member_named(user_name)
                .ok_or(format!("couldn't find member: {}", user_name))
                .map(Clone::clone)
        }
    }
}
